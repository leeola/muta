package muta

import (
	"io/ioutil"
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

func TestSrcStreamerNext(t *testing.T) {
	tmpDir := filepath.Join("_test", "fixtures")

	Convey("With no previous Streamers", t, func() {
		Convey("It should load its own files right away", func() {
			s := &SrcStreamer{
				Sources: []string{filepath.Join(tmpDir, "hello")},
			}
			fi, r, err := s.Next(nil, nil)
			So(err, ShouldBeNil)
			So(fi.Name(), ShouldEqual, "hello")
			b, _ := ioutil.ReadAll(r)
			So(string(b), ShouldEqual, "hello")
			r.Close()
		})
	})

	Convey("With previous Streamers", t, func() {
		Convey("the files should be loaded in order", func() {
			s := Stream{
				&MockStreamer{
					Files: []string{"foo", "bar"},
				},
				&MockStreamer{
					Files: []string{"baz", "bat"},
				},
				&SrcStreamer{
					Sources: []string{filepath.Join(tmpDir, "hello")},
				},
			}
			fi, r, err := s.Next(nil, nil)
			So(err, ShouldBeNil)
			So(fi.Name(), ShouldEqual, "foo")
			b, _ := ioutil.ReadAll(r)
			So(string(b), ShouldEqual, "foo content")

			fi, r, err = s.Next(nil, nil)
			So(err, ShouldBeNil)
			So(fi.Name(), ShouldEqual, "bar")
			b, _ = ioutil.ReadAll(r)
			So(string(b), ShouldEqual, "bar content")

			fi, r, err = s.Next(nil, nil)
			So(err, ShouldBeNil)
			So(fi.Name(), ShouldEqual, "baz")
			b, _ = ioutil.ReadAll(r)
			So(string(b), ShouldEqual, "baz content")

			fi, r, err = s.Next(nil, nil)
			So(err, ShouldBeNil)
			So(fi.Name(), ShouldEqual, "bat")
			b, _ = ioutil.ReadAll(r)
			So(string(b), ShouldEqual, "bat content")

			fi, r, err = s.Next(nil, nil)
			So(err, ShouldBeNil)
			So(fi.Name(), ShouldEqual, "hello")
			b, _ = ioutil.ReadAll(r)
			So(string(b), ShouldEqual, "hello")
			r.Close()
		})
	})

}
