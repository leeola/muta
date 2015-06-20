package muta

import "io"

// A FuncStreamer is a single Function implementation of a Streamer. Best
// used only for very simplistic Streamers that do not need to store any
// additional state values.
//
// Everything you would do in a normal Next() method, you do here as well.
// This just lets you write an inline Plugin.
type FuncStreamer func(FileInfo, io.ReadCloser) (FileInfo,
	io.ReadCloser, error)

func (f FuncStreamer) Next(fi FileInfo, rc io.ReadCloser) (FileInfo,
	io.ReadCloser, error) {
	return f(fi, rc)
}
