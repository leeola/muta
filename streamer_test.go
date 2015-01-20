package muta

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSrcStreamer(t *testing.T) {
	Convey("Should not modify incoming chunks", t, func() {
		s := SrcStreamer([]string{}, SrcOpts{})
		fi := &FileInfo{}
		b := []byte("chunk")
		rfi, rb, err := s(fi, b)
		So(err, ShouldBeNil)
		So(rfi, ShouldEqual, fi)
		So(rb, ShouldResemble, b)
		s(fi, b)
	})

	Convey("Should load the given file and", t, func() {
		Convey("Populate FileInfo with the file info", func() {
			s := SrcStreamer([]string{"_test/fixtures/hello.md"}, SrcOpts{})
			fi, _, err := s(nil, nil)
			So(err, ShouldBeNil)
			So(fi, ShouldResemble, &FileInfo{
				Name: "hello.md",
				Path: "_test/fixtures",
			})
			// flush for defer file close
			s(nil, nil)
		})

		Convey("Return chunks of the file", func() {
			s := SrcStreamer([]string{"_test/fixtures/hello.md"}, SrcOpts{
				ReadSize: 5,
			})
			_, b, err := s(nil, nil)
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte("hello"))
			// flush for defer file close
			s(nil, nil)
		})

		Convey("Return multiple chunks of the file", func() {
			s := SrcStreamer([]string{"_test/fixtures/hello.md"}, SrcOpts{
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
			s := SrcStreamer([]string{"_test/fixtures/hello.md"}, SrcOpts{
				ReadSize: 5,
			})
			s(nil, nil)
			fi, _, err := s(nil, nil)
			So(err, ShouldBeNil)
			So(fi, ShouldResemble, &FileInfo{
				Name: "hello.md",
				Path: "_test/fixtures",
			})
		})

		Convey("Return a nil chunk at EOF", func() {
			s := SrcStreamer([]string{"_test/fixtures/hello.md"}, SrcOpts{
				ReadSize: 5,
			})
			s(nil, nil)
			_, b, err := s(nil, nil)
			So(err, ShouldBeNil)
			So(b, ShouldBeNil)
		})

		Convey("Trim byte array to length of data", func() {
			s := SrcStreamer([]string{"_test/fixtures/hello.md"}, SrcOpts{
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
}
