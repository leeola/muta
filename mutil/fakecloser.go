package mutil

import (
	"bytes"
	"io"
	"strings"
)

// The fakeReadClose struct provides us with a way to implement a
// ReadCloser over a normal io.Reader. To use this, see the ByteCloser()
// and ReadCloser() funcs.
type fakeReadClose struct {
	Reader io.Reader
}

func (rc *fakeReadClose) Read(p []byte) (int, error) {
	return rc.Reader.Read(p)
}

func (rc *fakeReadClose) Close() error {
	// nothing to do
	return nil
}

// Take in a byte array, and return an io.ReadCloser compatible
// bytes.Reader.
func ByteCloser(b []byte) io.ReadCloser {
	return ReadCloser(bytes.NewReader(b))
}

// Take in a plain io.Reader, and return a new Struct with a fake Close
// method.
func ReadCloser(r io.Reader) io.ReadCloser {
	return &fakeReadClose{r}
}

// Take in a string, and return an io.ReadCloser compatible
// string.Reader.
func StringCloser(s string) io.ReadCloser {
	return ReadCloser(strings.NewReader(s))
}
