package muta

import (
	"errors"
	"io/ioutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMockStreamerNext(t *testing.T) {
	Convey("Should pass through all files", t, func() {
		s := MockStreamer{Files: []string{"bar"}}
		fi, rc, err := s.Next(NewFileInfo("foo"), nil)
		So(err, ShouldBeNil)
		So(rc, ShouldBeNil)
		So(fi, ShouldNotBeNil)
		So(fi.Name(), ShouldEqual, "foo")
	})

	Convey("Should create FileInfos", t, func() {
		s := MockStreamer{Files: []string{"foo"}}
		fi, _, err := s.Next(nil, nil)
		So(err, ShouldBeNil)
		So(fi, ShouldNotBeNil)
		So(fi.Name(), ShouldEqual, "foo")
	})

	Convey("Should create file content", t, func() {
		s := MockStreamer{Files: []string{"foo"}}
		_, rc, err := s.Next(nil, nil)
		So(err, ShouldBeNil)
		So(rc, ShouldNotBeNil)
		b, err := ioutil.ReadAll(rc)
		So(err, ShouldBeNil)
		So(string(b), ShouldEqual, "foo content")
	})

	Convey("Should create use provided content", t, func() {
		s := MockStreamer{
			Files:    []string{"foo"},
			Contents: []string{"bar"},
		}
		_, rc, err := s.Next(nil, nil)
		So(err, ShouldBeNil)
		So(rc, ShouldNotBeNil)
		b, err := ioutil.ReadAll(rc)
		So(err, ShouldBeNil)
		So(string(b), ShouldEqual, "bar")
	})

	Convey("Should use provided errors", t, func() {
		s := MockStreamer{
			Files:  []string{"foo"},
			Errors: []error{errors.New("Foo")},
		}
		_, _, err := s.Next(nil, nil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "Foo")
	})
}
