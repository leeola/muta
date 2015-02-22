package logging

import (
	"bytes"
	"io/ioutil"
	"testing"
)
import . "github.com/smartystreets/goconvey/convey"

func TestLoggerSetLevel(t *testing.T) {
	Convey("Should set the LogLevel", t, func() {
		l := NewLogger(ioutil.Discard)
		l.SetLevel(ERROR)
		So(l.logLevel, ShouldEqual, ERROR)
	})
}

func TestLoggerSetTags(t *testing.T) {
	Convey("Should set the tags", t, func() {
		l := NewLogger(ioutil.Discard)
		l.SetTags("foo")
		So(l.tags[0].String(), ShouldEqual, "^foo$")
	})

	Convey(`Should default to nil`, t, func() {
		l := NewLogger(ioutil.Discard)
		l.SetTags()
		So(l.tags, ShouldBeNil)
	})
}

func TestLoggerlog(t *testing.T) {
	Convey("Should log messages to the given writer", t, func() {
		var b bytes.Buffer
		l := &Logger{&b, INFO, nil}
		l.log(INFO, nil, "foo")
		l.log(INFO, nil, "bar")
		So(b.String(), ShouldEqual, "foo\nbar\n")
	})

	Convey("Should prepend the first tag to the log", t, func() {
		var b bytes.Buffer
		l := &Logger{&b, INFO, nil}
		l.log(INFO, []string{"foo", "bar"}, "a")
		So(b.String(), ShouldEqual, "[foo] a\n")
	})

	Convey("Should filter messages by tags", t, func() {
		var b bytes.Buffer
		l := &Logger{&b, INFO, nil}
		l.SetTags("b*")
		l.log(INFO, nil, "a")
		l.log(INFO, []string{"foo"}, "b")
		l.log(INFO, []string{"bar"}, "c")
		So(b.String(), ShouldEqual, "[bar] c\n")
	})

	Convey("Should filter messages by LogLevel", t, func() {
		var b bytes.Buffer
		l := NewLogger(&b)
		l.SetLevel(WARN)
		l.log(VERBOSE, nil, "a")
		l.log(DEBUG, nil, "b")
		l.log(INFO, nil, "c")
		l.log(WARN, nil, "d")
		l.log(ERROR, nil, "e")
		So(b.String(), ShouldEqual, "d\ne\n")
	})
}
