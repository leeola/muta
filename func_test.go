package muta

import (
	"io"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFuncStream(t *testing.T) {
	Convey("Should implement Streamer", t, func() {
		fn := func(FileInfo, io.ReadCloser) (FileInfo, io.ReadCloser, error) {
			return nil, nil, nil
		}
		var fs interface{} = FuncStreamer(fn)
		_, ok := fs.(Streamer)
		So(ok, ShouldBeTrue)
	})
}
