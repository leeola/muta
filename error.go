package muta

import (
	"io"
	"strings"
)

func NewErrorStreamer(msgs ...string) ErrorStreamer {
	return ErrorStreamer{Message: strings.Join(msgs, " ")}
}

// The ErrorStreamer fulfills both the Error and Streamer interfaces.
// No actual Streamer functionality is built in, and any usage of this
// Streamer returns itslf as an Error.
//
// This is useful for functions that return a Streamer, but may want to
// return an error.
//
type ErrorStreamer struct {
	Message string
}

func (s ErrorStreamer) Error() string {
	return s.Message
}

func (s ErrorStreamer) Next(FileInfo, io.ReadCloser) (FileInfo,
	io.ReadCloser, error) {

	return nil, nil, s
}
