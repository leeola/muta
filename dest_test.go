package muta

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDest(t *testing.T) {
	tmpDir := filepath.Join("_test", "tmp")

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

	os.Remove(filepath.Join(tmpDir, "file"))
	os.Remove(filepath.Join(tmpDir, "different_file"))

	Convey("Should write to the given file even if the filename "+
		"changes after opening the writer", t, func() {
		s := Dest(tmpDir)
		f := NewFileInfo("./file")
		_, _, err := s(f, []byte("foo"))
		So(err, ShouldBeNil)
		f.Name = "different_file"
		_, _, err = s(f, []byte("bar"))
		So(err, ShouldBeNil)
		_, _, err = s(nil, nil)
		So(err, ShouldBeNil)

		b, err := ioutil.ReadFile(filepath.Join(tmpDir, "file"))
		So(err, ShouldBeNil)
		So(b, ShouldResemble, []byte("foobar"))
		b, err = ioutil.ReadFile(filepath.Join(tmpDir, "different_file"))
		So(err, ShouldNotBeNil)
		So(b, ShouldBeNil)
	})
}
