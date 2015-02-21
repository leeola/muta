package muta

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSrcStreamer(t *testing.T) {
	Convey("Should pipe incoming chunks", t, func() {
		s := SrcStreamer([]string{}, SrcOpts{}).Stream
		fi := &FileInfo{}
		b := []byte("chunk")
		rfi, rb, err := s(fi, b)
		So(err, ShouldBeNil)
		So(rfi, ShouldEqual, fi)
		So(rb, ShouldResemble, b)
		s(fi, b)
	})

	Convey("Should return an error if the file cannot be found", t, func() {
		s := SrcStreamer([]string{"_test/fixtures/404"}, SrcOpts{}).Stream
		_, _, err := s(nil, nil)
		So(err, ShouldNotBeNil)
	})

	Convey("Should load the given file and", t, func() {
		Convey("Populate FileInfo with the file info", func() {
			s := SrcStreamer([]string{"_test/fixtures/hello"},
				SrcOpts{}).Stream
			fi, _, err := s(nil, nil)
			So(err, ShouldBeNil)
			So(fi, ShouldResemble, &FileInfo{
				Name:         "hello",
				Path:         "_test/fixtures",
				OriginalName: "hello",
				OriginalPath: "_test/fixtures",
				Ctx:          map[string]interface{}{},
			})
			// flush for defer file close
			s(nil, nil)
		})

		Convey("Return chunks of the file", func() {
			s := SrcStreamer([]string{"_test/fixtures/hello"}, SrcOpts{
				ReadSize: 5,
			}).Stream
			_, b, err := s(nil, nil)
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte("hello"))
			// flush for defer file close
			s(nil, nil)
		})

		Convey("Return multiple chunks of the file", func() {
			s := SrcStreamer([]string{"_test/fixtures/hello"}, SrcOpts{
				ReadSize: 3,
			}).Stream
			_, b, err := s(nil, nil)
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte("hel"))
			_, b, err = s(nil, nil)
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte("lo"))
			// flush for defer file close
			s(nil, nil)
		})

		Convey("Return a valid FileInfo at EOF", func() {
			s := SrcStreamer([]string{"_test/fixtures/hello"}, SrcOpts{
				ReadSize: 5,
			}).Stream
			s(nil, nil)
			fi, _, err := s(nil, nil)
			So(err, ShouldBeNil)
			So(fi, ShouldResemble, &FileInfo{
				Name:         "hello",
				Path:         "_test/fixtures",
				OriginalName: "hello",
				OriginalPath: "_test/fixtures",
				Ctx:          map[string]interface{}{},
			})
		})

		Convey("Return a nil chunk at EOF", func() {
			s := SrcStreamer([]string{"_test/fixtures/hello"}, SrcOpts{
				ReadSize: 5,
			}).Stream
			s(nil, nil)
			_, b, err := s(nil, nil)
			So(err, ShouldBeNil)
			So(b, ShouldBeNil)
		})

		Convey("Trim byte array to length of data", func() {
			s := SrcStreamer([]string{"_test/fixtures/hello"}, SrcOpts{
				ReadSize: 4,
			}).Stream
			_, b, _ := s(nil, nil)
			So(b, ShouldResemble, []byte("hell"))
			_, b, _ = s(nil, nil)
			So(b, ShouldResemble, []byte("o"))
			// flush for defer file close
			s(nil, nil)
		})
	})

	Convey("Should stream any number of files", t, func() {
		s := SrcStreamer([]string{
			"_test/fixtures/hello",
			"_test/fixtures/world",
		}, SrcOpts{}).Stream
		helloFi := &FileInfo{
			Name:         "hello",
			Path:         "_test/fixtures",
			OriginalName: "hello",
			OriginalPath: "_test/fixtures",
			Ctx:          map[string]interface{}{},
		}
		worldFi := &FileInfo{
			Name:         "world",
			Path:         "_test/fixtures",
			OriginalName: "world",
			OriginalPath: "_test/fixtures",
			Ctx:          map[string]interface{}{},
		}
		fi, chunk, err := s(nil, nil)
		So(err, ShouldBeNil)
		So(fi, ShouldResemble, helloFi)
		So(chunk, ShouldResemble, []byte("hello"))

		fi, chunk, err = s(nil, nil)
		So(err, ShouldBeNil)
		So(fi, ShouldResemble, helloFi)
		So(chunk, ShouldBeNil) // EOF

		fi, chunk, err = s(nil, nil)
		So(err, ShouldBeNil)
		So(fi, ShouldResemble, worldFi)
		So(chunk, ShouldResemble, []byte("world"))

		fi, chunk, err = s(nil, nil)
		So(err, ShouldBeNil)
		So(fi, ShouldResemble, worldFi)
		So(chunk, ShouldBeNil) // EOF

		fi, chunk, err = s(nil, nil)
		So(err, ShouldBeNil)
		So(fi, ShouldBeNil) // EOS
	})

	Convey("Should support globbing", t, func() {
		s := SrcStreamer([]string{"_test/fixtures/*.md"}, SrcOpts{}).Stream
		files := []string{}
		var err error
		for true {
			fi, chunk, serr := s(nil, nil)
			err = serr
			if err != nil || fi == nil {
				break
			}
			if chunk == nil {
				// Only add the file when the Streamer signals EOF
				files = append(files, fi.Name)
			}
		}
		So(err, ShouldBeNil)
		So(files, ShouldResemble, []string{"hello.md", "world.md"})
	})

	Convey("Should instantiate the Ctx map", t, func() {
		s := SrcStreamer([]string{"_test/fixtures/hello"}, SrcOpts{}).Stream
		fi, _, err := s(nil, nil)
		So(err, ShouldBeNil)
		So(fi.Ctx, ShouldNotBeNil)
	})
}
