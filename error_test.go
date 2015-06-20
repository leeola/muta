package muta

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestErrorStreamerNext(t *testing.T) {
	Convey("Should return itself as an error", t, func() {
		s := ErrorStreamer{Message: "err msg"}
		var sErr error = s
		fi, rc, err := s.Next(nil, nil)
		So(fi, ShouldBeNil)
		So(rc, ShouldBeNil)
		So(err.Error(), ShouldEqual, sErr.Error())
	})
}
