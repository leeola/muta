package muta

import (
	"io/ioutil"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSrcStreamer(t *testing.T) {
	Convey("Should pipe incoming chunks", t, func() {
		s := SrcStreamer([]string{}, SrcOpts{})
		fi := &FileInfo{}
		b := []byte("chunk")
		rfi, rb, err := s(fi, b)
		So(err, ShouldBeNil)
		So(rfi, ShouldEqual, fi)
		So(rb, ShouldResemble, b)
		s(fi, b)
	})

	Convey("Should return an error if the file cannot be found", t, func() {
		s := SrcStreamer([]string{"_test/fixtures/404"}, SrcOpts{})
		_, _, err := s(nil, nil)
		So(err, ShouldNotBeNil)
	})

	Convey("Should load the given file and", t, func() {
		Convey("Populate FileInfo with the file info", func() {
			s := SrcStreamer([]string{"_test/fixtures/hello"}, SrcOpts{})
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
			})
			_, b, err := s(nil, nil)
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte("hello"))
			// flush for defer file close
			s(nil, nil)
		})

		Convey("Return multiple chunks of the file", func() {
			s := SrcStreamer([]string{"_test/fixtures/hello"}, SrcOpts{
				ReadSize: 3,
			})
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
			})
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
			})
			s(nil, nil)
			_, b, err := s(nil, nil)
			So(err, ShouldBeNil)
			So(b, ShouldBeNil)
		})

		Convey("Trim byte array to length of data", func() {
			s := SrcStreamer([]string{"_test/fixtures/hello"}, SrcOpts{
				ReadSize: 4,
			})
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
		}, SrcOpts{})
		fi, chunk, err := s(nil, nil)
		So(err, ShouldBeNil)
		So(fi, ShouldResemble, &FileInfo{
			Name:         "hello",
			Path:         "_test/fixtures",
			OriginalName: "hello",
			OriginalPath: "_test/fixtures",
			Ctx:          map[string]interface{}{},
		})
		So(chunk, ShouldResemble, []byte("hello"))
		fi, chunk, err = s(nil, nil)
		So(err, ShouldBeNil)
		So(fi, ShouldResemble, &FileInfo{
			Name:         "hello",
			Path:         "_test/fixtures",
			OriginalName: "hello",
			OriginalPath: "_test/fixtures",
			Ctx:          map[string]interface{}{},
		})
		So(chunk, ShouldBeNil) // EOF
		fi, chunk, err = s(nil, nil)
		So(err, ShouldBeNil)
		So(fi, ShouldResemble, &FileInfo{
			Name:         "world",
			Path:         "_test/fixtures",
			OriginalName: "world",
			OriginalPath: "_test/fixtures",
			Ctx:          map[string]interface{}{},
		})
		So(chunk, ShouldResemble, []byte("world"))
		fi, chunk, err = s(nil, nil)
		So(err, ShouldBeNil)
		So(fi, ShouldResemble, &FileInfo{
			Name:         "world",
			Path:         "_test/fixtures",
			OriginalName: "world",
			OriginalPath: "_test/fixtures",
			Ctx:          map[string]interface{}{},
		})
		So(chunk, ShouldBeNil) // EOF
		fi, chunk, err = s(nil, nil)
		So(err, ShouldBeNil)
		So(fi, ShouldBeNil) // EOS
	})

	Convey("Should support globbing", t, func() {
		s := SrcStreamer([]string{"_test/fixtures/*.md"}, SrcOpts{})
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
		s := SrcStreamer([]string{"_test/fixtures/hello"}, SrcOpts{})
		fi, _, err := s(nil, nil)
		So(err, ShouldBeNil)
		So(fi.Ctx, ShouldNotBeNil)
	})
}

func TestDest(t *testing.T) {
	os.RemoveAll("_test/tmp/dest")

	Convey("Should create the the destination if needed", t, func() {
		s := Dest("_test/tmp/dest")
		f := &FileInfo{
			Name: "file",
			Path: ".",
		}
		c := []byte("chunk")
		s(f, c)
		s(nil, nil)
		osFi, err := os.Stat("_test/tmp/dest")
		So(err, ShouldBeNil)
		So(osFi.IsDir(), ShouldBeTrue)
	})

	os.RemoveAll("_test/tmp/path")

	Convey("Should create the path in the dest if needed", t, func() {
		s := Dest("_test/tmp")
		f := &FileInfo{
			Name: "file",
			Path: "path/foo/bar",
		}
		c := []byte("chunk")
		_, _, err := s(f, c)
		So(err, ShouldBeNil)
		osFi, err := os.Stat("_test/tmp/path/foo/bar")
		So(err, ShouldBeNil)
		So(osFi.IsDir(), ShouldBeTrue)
	})

	os.Remove("_test/tmp/file")

	Convey("Should create the file in the destination", t, func() {
		s := Dest("_test/tmp")
		f := &FileInfo{
			Name: "file",
			Path: ".",
		}
		c := []byte("foo")
		_, _, err := s(f, c)
		So(err, ShouldBeNil)
		// Signal EOF
		_, _, err = s(nil, nil)
		So(err, ShouldBeNil)
		// Test the file
		_, err = os.Stat("_test/tmp/file")
		So(err, ShouldBeNil)
	})

	os.Remove("_test/tmp/file")

	Convey("Should write incoming bytes to the given file", t, func() {
		s := Dest("_test/tmp")
		f := &FileInfo{
			Name: "file",
			Path: ".",
		}
		_, _, err := s(f, []byte("foo"))
		So(err, ShouldBeNil)
		_, _, err = s(f, []byte("bar"))
		So(err, ShouldBeNil)
		_, _, err = s(f, []byte("baz"))
		So(err, ShouldBeNil)
		_, _, err = s(nil, nil)
		So(err, ShouldBeNil)

		b, err := ioutil.ReadFile("_test/tmp/file")
		So(err, ShouldBeNil)
		So(b, ShouldResemble, []byte("foobarbaz"))
	})

	Convey("Should not allow writing outside of the destination", t, nil)
}
