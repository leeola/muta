package muta

import (
	"path/filepath"
	"testing"

	"github.com/leeola/muta/logging"
	. "github.com/smartystreets/goconvey/convey"
)

func init() {
	logging.SetLevel(logging.ERROR)
}

func TestGlobsToBase(t *testing.T) {
	Convey("Should return the correct base", t, func() {
		So(globsToBase("."), ShouldEqual, ".")
		So(globsToBase("foo/bar.baz"), ShouldEqual, "foo")
		So(globsToBase("foo/*.baz"), ShouldEqual, "foo")
		So(globsToBase("foo/bar/**/*.baz"), ShouldEqual, "foo/bar")
		So(globsToBase(
			"foo/bar/baz",
			"foo/**/baz",
		), ShouldEqual, "foo")
	})
}

func TestSrcStreamer(t *testing.T) {
	tmpDir := filepath.Join("_test", "fixtures")

	Convey("Should pipe incoming chunks", t, func() {
		s := NewSrcStreamer().Stream
		fi := &FileInfo{}
		b := []byte("chunk")
		rfi, rb, err := s(fi, b)
		So(err, ShouldBeNil)
		So(rfi, ShouldEqual, fi)
		So(rb, ShouldResemble, b)
		s(fi, b)
	})

	Convey("Should return an error if the file cannot be found", t, func() {
		s := NewSrcStreamer("_test/fixtures/404").Stream
		_, _, err := s(nil, nil)
		So(err, ShouldNotBeNil)
	})

	Convey("Should load the given file and", t, func() {
		Convey("Populate FileInfo with the file info", func() {
			s := NewSrcStreamer("_test/fixtures/hello").Stream
			fi, _, err := s(nil, nil)
			So(err, ShouldBeNil)
			So(fi, ShouldResemble, &FileInfo{
				Name:         "hello",
				Path:         ".",
				OriginalName: "hello",
				OriginalPath: "_test/fixtures",
				Ctx:          map[string]interface{}{},
			})
			// flush for defer file close
			s(nil, nil)
		})

		Convey("Return chunks of the file", func() {
			s := (&SrcStreamer{
				Sources:  []string{"_test/fixtures/hello"},
				ReadSize: 5,
			}).Init().Stream
			_, b, err := s(nil, nil)
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte("hello"))
			// flush for defer file close
			s(nil, nil)
		})

		Convey("Return multiple chunks of the file", func() {
			s := (&SrcStreamer{
				Sources:  []string{"_test/fixtures/hello"},
				ReadSize: 3,
			}).Init().Stream
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
			s := (&SrcStreamer{
				Sources:  []string{"_test/fixtures/hello"},
				ReadSize: 5,
			}).Init().Stream
			s(nil, nil)
			fi, _, err := s(nil, nil)
			So(err, ShouldBeNil)
			So(fi, ShouldResemble, &FileInfo{
				Name:         "hello",
				Path:         ".",
				OriginalName: "hello",
				OriginalPath: "_test/fixtures",
				Ctx:          map[string]interface{}{},
			})
		})

		Convey("Return a nil chunk at EOF", func() {
			s := (&SrcStreamer{
				Sources:  []string{"_test/fixtures/hello"},
				ReadSize: 5,
			}).Init().Stream
			s(nil, nil)
			_, b, err := s(nil, nil)
			So(err, ShouldBeNil)
			So(b, ShouldBeNil)
		})

		Convey("Trim byte array to length of data", func() {
			s := (&SrcStreamer{
				Sources:  []string{"_test/fixtures/hello"},
				ReadSize: 4,
			}).Init().Stream
			_, b, _ := s(nil, nil)
			So(b, ShouldResemble, []byte("hell"))
			_, b, _ = s(nil, nil)
			So(b, ShouldResemble, []byte("o"))
			// flush for defer file close
			s(nil, nil)
		})
	})

	Convey("Should stream any number of files", t, func() {
		s := NewSrcStreamer(
			"_test/fixtures/hello",
			"_test/fixtures/world",
		).Stream
		helloFi := &FileInfo{
			Name:         "hello",
			Path:         ".",
			OriginalName: "hello",
			OriginalPath: "_test/fixtures",
			Ctx:          map[string]interface{}{},
		}
		worldFi := &FileInfo{
			Name:         "world",
			Path:         ".",
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
		s := NewSrcStreamer("_test/fixtures/*.md").Stream
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
		s := NewSrcStreamer("_test/fixtures/hello").Stream
		fi, _, err := s(nil, nil)
		So(err, ShouldBeNil)
		So(fi.Ctx, ShouldNotBeNil)
	})

	Convey("Should trim the base path", t, func() {
		Convey("Up to the first glob", func() {
			p := []string{filepath.Join(tmpDir, "nested", "markdown", "*.md")}
			s := NewSrcStreamer(p...).Stream
			var count int
			var lastFi *FileInfo
			for fi, _, err := s(nil, nil); fi != nil; fi, _, err = s(nil, nil) {
				if fi == lastFi {
					continue
				}
				So(err, ShouldBeNil)
				So(fi.Path, ShouldEqual, ".")
				lastFi = fi
				count++
			}
			So(count, ShouldNotEqual, 0)
		})
	})
}
