package muta

import (
	"io"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFuncStream(t *testing.T) {
	Convey("Should set the funcStreamer", t, func() {

		fn := func(s Streamer) (*FileInfo, io.ReadCloser, error) {
			return nil, nil, nil
		}
		s := FuncStream(fn).Embed(nil)
		fns, _ := s.(*funcStream)
		So(fns.funcStreamer, ShouldEqual, fn)
	})
}

func TestFuncStreamNext(t *testing.T) {
	Convey("Should call the funcStreamer", t, func() {
		called := false
		s := &funcStream{
			funcStreamer: func(s Streamer) (*FileInfo, io.ReadCloser, error) {
				called = true
				return &FileInfo{Name: "foo"}, nil, nil
			}}
		fi, _, _ := s.Next()
		So(called, ShouldBeTrue)
		So(fi.Name, ShouldEqual, "foo")
	})
}
