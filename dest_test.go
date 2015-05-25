package muta

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)
import (
	//"github.com/leeola/muta/mtesting"
	"github.com/leeola/muta/logging"
	. "github.com/smartystreets/goconvey/convey"
)

func init() {
	logging.SetLevel(logging.ERROR)
}

func TestDestWithOpts(t *testing.T) {
	tmpDir := filepath.Join("_test", "tmp")

	os.RemoveAll(filepath.Join(tmpDir, "dest"))

	Convey("Should create the destination if needed", t, func() {
		DestWithOpts(filepath.Join(tmpDir, "dest"), DestOpts{
			Clean: false, Overwrite: false}).Embed(nil)
		osFi, err := os.Stat(filepath.Join(tmpDir, "dest"))
		So(err, ShouldBeNil)
		So(osFi.IsDir(), ShouldBeTrue)
	})

	ioutil.WriteFile(filepath.Join(tmpDir, "dest", "file"),
		[]byte("REMOVE ME"), 0777)

	Convey("Should remove the destination if Clean is true", t, func() {
		DestWithOpts(filepath.Join(tmpDir, "dest"), DestOpts{
			Clean: true, Overwrite: false}).Embed(nil)
		osFi, err := os.Stat(filepath.Join(tmpDir, "dest"))
		So(err, ShouldBeNil)
		So(osFi.IsDir(), ShouldBeTrue)
		// Now ensure that this is a new dir, by checking for the prev file
		_, err = os.Stat(filepath.Join(tmpDir, "dest", "file"))
		So(err, ShouldNotBeNil)
		So(os.IsNotExist(err), ShouldBeTrue)
	})

	ioutil.WriteFile(filepath.Join(tmpDir, "dest", "file"),
		[]byte("REMOVE ME"), 0777)

	Convey("Should not remove the destination if Clean isnt set", t, func() {
		DestWithOpts(filepath.Join(tmpDir, "dest"), DestOpts{
			Clean: false, Overwrite: false}).Embed(nil)
		osFi, err := os.Stat(filepath.Join(tmpDir, "dest"))
		So(err, ShouldBeNil)
		So(osFi.IsDir(), ShouldBeTrue)
		// Now ensure that this isnt a new dir, by checking for the prev file
		osFi, err = os.Stat(filepath.Join(tmpDir, "dest", "file"))
		So(err, ShouldBeNil)
		So(osFi.IsDir(), ShouldBeFalse)
	})
}

func TestDestStreamerNext(t *testing.T) {
	tmpDir := filepath.Join("_test", "tmp")

	os.RemoveAll(filepath.Join(tmpDir, "file"))

	Convey("Should create the file in the destination", t, func() {
		s := &DestStreamer{
			Streamer:    &MockStreamer{Files: []string{"file"}},
			Destination: tmpDir,
			Opts: DestOpts{
				Clean: false, Overwrite: false},
		}
		_, _, err := s.Next()
		So(err, ShouldBeNil)
		osFi, err := os.Stat(filepath.Join(tmpDir, "dest", "file"))
		So(err, ShouldBeNil)
		So(osFi.IsDir(), ShouldBeFalse)
	})

	os.RemoveAll(filepath.Join(tmpDir, "file"))

	Convey("Should write all Read data to the file", t, func() {
		s := &DestStreamer{
			Streamer:    &MockStreamer{Files: []string{"file"}},
			Destination: tmpDir,
			Opts:        DestOpts{Clean: false, Overwrite: false},
		}
		_, _, err := s.Next()
		So(err, ShouldBeNil)
		b, err := ioutil.ReadFile(filepath.Join(tmpDir, "file"))
		So(err, ShouldBeNil)
		So(string(b), ShouldEqual, "file content")
	})

	os.RemoveAll(filepath.Join(tmpDir, "file1"))
	os.RemoveAll(filepath.Join(tmpDir, "file2"))

	Convey("Should write all Read data to the file, repeatedly", t, func() {
		s := &DestStreamer{
			Streamer:    &MockStreamer{Files: []string{"file1", "file2"}},
			Destination: tmpDir,
			Opts:        DestOpts{Clean: false, Overwrite: false},
		}
		_, _, err := s.Next()
		So(err, ShouldBeNil)
		b, err := ioutil.ReadFile(filepath.Join(tmpDir, "file1"))
		So(err, ShouldBeNil)
		So(string(b), ShouldEqual, "file1 content")

		_, _, err = s.Next()
		So(err, ShouldBeNil)
		b, err = ioutil.ReadFile(filepath.Join(tmpDir, "file2"))
		So(err, ShouldBeNil)
		So(string(b), ShouldEqual, "file2 content")
	})

	os.RemoveAll(filepath.Join(tmpDir, "file"))
	ioutil.WriteFile(filepath.Join(tmpDir, "file"),
		[]byte("REPLACE ME"), 0777)

	Convey("Should overwrite all Read data to the file, if set", t, func() {
		s := &DestStreamer{
			Streamer:    &MockStreamer{Files: []string{"file"}},
			Destination: tmpDir,
			Opts:        DestOpts{Clean: false, Overwrite: true},
		}
		_, _, err := s.Next()
		So(err, ShouldBeNil)
		b, err := ioutil.ReadFile(filepath.Join(tmpDir, "file"))
		So(err, ShouldBeNil)
		So(string(b), ShouldEqual, "file content")
	})

	os.RemoveAll(filepath.Join(tmpDir, "file"))
	ioutil.WriteFile(filepath.Join(tmpDir, "file"),
		[]byte("DON'T REPLACE ME"), 0777)

	Convey("Should not overwrite all Read data to the file, if not set", t, func() {
		s := &DestStreamer{
			Streamer:    &MockStreamer{Files: []string{"file"}},
			Destination: tmpDir,
			Opts:        DestOpts{Clean: false, Overwrite: false},
		}
		_, _, err := s.Next()
		So(err, ShouldNotBeNil)
		b, err := ioutil.ReadFile(filepath.Join(tmpDir, "file"))
		So(err, ShouldBeNil)
		So(string(b), ShouldEqual, "DON'T REPLACE ME")
	})

	os.RemoveAll(filepath.Join(tmpDir, "path"))

	Convey("Should create the file path in the destination", t, func() {
		s := &DestStreamer{
			Streamer: &MockStreamer{
				Files: []string{filepath.Join("path", "path_file")}},
			Destination: tmpDir,
			Opts:        DestOpts{Clean: false, Overwrite: false},
		}
		_, _, err := s.Next()
		So(err, ShouldBeNil)

		osFi, err := os.Stat(filepath.Join(tmpDir, "path"))
		So(err, ShouldBeNil)
		So(osFi.IsDir(), ShouldBeTrue)

		b, err := ioutil.ReadFile(filepath.Join(tmpDir, "path", "path_file"))
		So(err, ShouldBeNil)
		So(string(b), ShouldEqual, fmt.Sprintf("%s content",
			filepath.Join("path", "path_file")))
	})

	Convey("Should not allow writing outside of the destination", t, nil)

}
