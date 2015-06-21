package muta

import (
	"errors"
	"io"
	"io/ioutil"
	"testing"

	"github.com/leeola/muta/mutil"

	. "github.com/smartystreets/goconvey/convey"
)

func TestStreamPipe(t *testing.T) {
	Convey("Should append a Streamer", t, func() {
		a := &MockStreamer{}
		b := &MockStreamer{}
		s := Stream{}
		s = s.Pipe(a).Pipe(b)
		So(len(s), ShouldEqual, 2)
		So(s[0], ShouldEqual, a)
		So(s[1], ShouldEqual, b)
	})

	Convey("When an Error Streamer is Piped, the Stream Should", t, func() {
		Convey("resize Stream to only the Error", func() {
			a := &MockStreamer{}
			err := &ErrorStreamer{}
			s := Stream{}
			s = s.Pipe(a).Pipe(err)
			So(len(s), ShouldEqual, 1)
			So(s[0], ShouldEqual, err)
		})
	})
}

func TestStreamNextFrom(t *testing.T) {
	Convey("Should return Streamer data", t, func() {
		s := Stream{&MockStreamer{
			Files:  []string{"foo", "bar"},
			Errors: []error{errors.New("Foo")},
		}}

		fi, rc, err := s.NextFrom(0, nil, nil)
		So(fi, ShouldNotBeNil)
		So(fi.Name(), ShouldEqual, "foo")

		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "Foo")

		So(rc, ShouldNotBeNil)
		b, err := ioutil.ReadAll(rc)
		So(err, ShouldBeNil)
		So(string(b), ShouldEqual, "foo content")

		fi, rc, err = s.NextFrom(0, nil, nil)
		So(err, ShouldBeNil)

		So(fi, ShouldNotBeNil)
		So(fi.Name(), ShouldEqual, "bar")

		b, err = ioutil.ReadAll(rc)
		So(err, ShouldBeNil)
		So(string(b), ShouldEqual, "bar content")
	})

	Convey("Should call Streamers from the given index", t, func() {
		s := Stream{
			&MockStreamer{
				Files: []string{"foo"},
			},
			&MockStreamer{
				Files: []string{"bar"},
			},
		}

		fi, rc, _ := s.NextFrom(1, nil, nil)
		So(fi, ShouldNotBeNil)
		So(fi.Name(), ShouldEqual, "bar")

		So(rc, ShouldNotBeNil)
		b, err := ioutil.ReadAll(rc)
		So(err, ShouldBeNil)
		So(string(b), ShouldEqual, "bar content")
	})

	Convey("Should stop calling on error", t, func() {
		callCount := 0
		s := Stream{
			FuncStreamer(func(fi FileInfo, rc io.ReadCloser) (
				FileInfo, io.ReadCloser, error) {
				callCount++
				return NewFileInfo("foo"), mutil.StringCloser("foo"),
					errors.New("foo")
			}),
			FuncStreamer(func(fi FileInfo, rc io.ReadCloser) (
				FileInfo, io.ReadCloser, error) {
				callCount++
				return NewFileInfo("bar"), mutil.StringCloser("bar"), nil
			}),
		}

		_, _, err := s.NextFrom(0, nil, nil)
		So(err, ShouldNotBeNil)
		So(callCount, ShouldEqual, 1)
	})

	Convey("Should not call subsequent Streamers if no file is returned", t, func() {

		callCount := 0
		s := Stream{
			FuncStreamer(func(fi FileInfo, rc io.ReadCloser) (
				FileInfo, io.ReadCloser, error) {
				callCount++
				return nil, nil, nil
			}),
			FuncStreamer(func(fi FileInfo, rc io.ReadCloser) (
				FileInfo, io.ReadCloser, error) {
				callCount++
				return NewFileInfo("bar"), mutil.StringCloser("bar"), nil
			}),
		}

		s.NextFrom(0, nil, nil)
		So(callCount, ShouldEqual, 1)
	})

	Convey("Should feed return each Streamer into the next", t, func() {
		var srcFi, retFi FileInfo
		var srcRc, retRc io.ReadCloser
		srcFi, srcRc = NewFileInfo("bar"), mutil.StringCloser("bar")

		s := Stream{
			FuncStreamer(func(fi FileInfo, rc io.ReadCloser) (
				FileInfo, io.ReadCloser, error) {
				return srcFi, srcRc, nil
			}),
			FuncStreamer(func(fi FileInfo, rc io.ReadCloser) (
				FileInfo, io.ReadCloser, error) {
				retFi, retRc = fi, rc
				return nil, nil, nil
			}),
		}

		s.NextFrom(0, nil, nil)
		So(retFi, ShouldEqual, srcFi)
		So(retRc, ShouldEqual, srcRc)
	})
}

func TestStreamStream(t *testing.T) {
	Convey("Should call Next() until all Streamers return nil", t, func() {
		aCallCount := 0
		bCallCount := 0
		s := Stream{
			FuncStreamer(func(_ FileInfo, _ io.ReadCloser) (
				FileInfo, io.ReadCloser, error) {
				aCallCount++
				if aCallCount <= 2 {
					return NewFileInfo("foo"), nil, nil
				}
				return nil, nil, nil
			}),
			FuncStreamer(func(fi FileInfo, _ io.ReadCloser) (
				FileInfo, io.ReadCloser, error) {
				bCallCount++
				if fi == nil && bCallCount <= 4 {
					return NewFileInfo("bar"), nil, nil
				}
				return nil, nil, nil
			}),
		}

		err := s.Stream()
		So(err, ShouldBeNil)
		So(aCallCount, ShouldEqual, 3)
		So(bCallCount, ShouldEqual, 5)
	})

	Convey("Should stop calling Next() if an error is returned", t, func() {
		callCount := 0
		s := Stream{
			FuncStreamer(func(fi FileInfo, rc io.ReadCloser) (
				FileInfo, io.ReadCloser, error) {
				callCount++
				return NewFileInfo("foo"), rc, errors.New("Foo")
			}),
		}
		err := s.Stream()
		So(err, ShouldNotBeNil)
		So(callCount, ShouldEqual, 1)
	})

	Convey("Should not pass a Streamer it's own return values", t, func() {
		var argFi FileInfo
		var stop bool
		retFi := NewFileInfo("foo")
		s := Stream{
			FuncStreamer(func(fi FileInfo, _ io.ReadCloser) (
				FileInfo, io.ReadCloser, error) {
				argFi = fi
				if stop {
					return nil, nil, nil
				}
				stop = true
				return retFi, mutil.StringCloser("foo"), nil
			}),
		}

		s.Stream()
		So(argFi, ShouldNotEqual, retFi)
	})
}
