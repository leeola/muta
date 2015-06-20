package muta

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"testing"

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

func TestStreamNext(t *testing.T) {
	sa := FuncStreamer(func(fi FileInfo, rc io.ReadCloser) (
		FileInfo, io.ReadCloser, error) {
		fi = NewFileInfo(fmt.Sprintf("%sa", fi.Name()))
		return fi, rc, nil
	})
	sb := FuncStreamer(func(fi FileInfo, rc io.ReadCloser) (
		FileInfo, io.ReadCloser, error) {
		fi = NewFileInfo(fmt.Sprintf("%sb", fi.Name()))
		return fi, rc, nil
	})

	Convey("Should pipe data to all Streamers", t, func() {
		s := Stream{sa, sb}

		fi, rc, err := s.Next(NewFileInfo("foo"), nil)
		So(err, ShouldBeNil)
		So(rc, ShouldBeNil)
		So(fi.Name(), ShouldEqual, "fooab")
	})

	Convey("Should stop piping data on Streamer error", t, func() {
		s := Stream{
			&MockStreamer{
				Files:  []string{"foo"},
				Errors: []error{errors.New("Foo")},
			},
			sa,
		}

		fi, rc, err := s.Next(nil, nil)
		So(err, ShouldNotBeNil)
		So(rc, ShouldNotBeNil)
		So(fi.Name(), ShouldEqual, "foo")
	})

	Convey("Should return Streamers data", t, func() {
		s := Stream{
			&MockStreamer{
				Files:  []string{"foo", "bar"},
				Errors: []error{nil, errors.New("Bar")},
			},
			&MockStreamer{
				Files:  []string{"baz", "bat"},
				Errors: []error{nil, errors.New("Bat")},
			},
		}

		fi, rc, err := s.Next(nil, nil)
		So(err, ShouldBeNil)
		So(fi, ShouldNotBeNil)
		So(fi.Name(), ShouldEqual, "foo")
		b, err := ioutil.ReadAll(rc)
		So(err, ShouldBeNil)
		So(string(b), ShouldEqual, "foo content")

		fi, rc, err = s.Next(nil, nil)
		So(err, ShouldNotBeNil)
		So(fi, ShouldNotBeNil)
		So(fi.Name(), ShouldEqual, "bar")
		b, err = ioutil.ReadAll(rc)
		So(err, ShouldBeNil)
		So(string(b), ShouldEqual, "bar content")

		fi, rc, err = s.Next(nil, nil)
		So(err, ShouldBeNil)
		So(fi, ShouldNotBeNil)
		So(fi.Name(), ShouldEqual, "baz")
		b, err = ioutil.ReadAll(rc)
		So(err, ShouldBeNil)
		So(string(b), ShouldEqual, "baz content")

		fi, rc, err = s.Next(nil, nil)
		So(err, ShouldNotBeNil)
		So(fi, ShouldNotBeNil)
		So(fi.Name(), ShouldEqual, "bat")
		b, err = ioutil.ReadAll(rc)
		So(err, ShouldBeNil)
		So(string(b), ShouldEqual, "bat content")
	})
}

func TestStreamStream(t *testing.T) {
	Convey("Should call Next() until all Streamers return nil", t, func() {
		aCallCount := 0
		bCallCount := 0
		s := Stream{
			FuncStreamer(func(fi FileInfo, rc io.ReadCloser) (
				FileInfo, io.ReadCloser, error) {
				aCallCount++
				if aCallCount <= 2 {
					fi = NewFileInfo("foo")
				}
				return fi, rc, nil
			}),
			FuncStreamer(func(fi FileInfo, rc io.ReadCloser) (
				FileInfo, io.ReadCloser, error) {
				bCallCount++
				if fi == nil && bCallCount <= 4 {
					fi = NewFileInfo("bar")
				}
				return fi, rc, nil
			}),
		}

		err := s.Stream()
		So(err, ShouldBeNil)
		So(aCallCount, ShouldEqual, 5)
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
}
