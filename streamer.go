package muta

import (
	"io"
	"path/filepath"
	"strings"
)

// Streamer implements the Next() method, which will be repeatedly called
// with FileInfo and ReadClosers from previous Streamers.
//
// FileInfo can be embedded in structs with additional methods, to provide
// additional functionality for a given file.
type Streamer interface {
	Next(FileInfo, io.ReadCloser) (FileInfo, io.ReadCloser, error)
}

// FileInfo is the base interface for getting and setting file info
// for the given file.
type FileInfo interface {
	// A getter and setter for the Name of the file object.
	Name() string
	SetName(string)

	// A getter and setter for the path of the file.
	Path() string
	SetPath(string)

	// A getter for the unmodified name and path of the file, as it was
	// at creation time.
	OriginalName() string
	OriginalPath() string

	// Ctx is a map[string]interface{} available for simple data storage
	// and/or passing. Generally it is recommended not to use it, and
	// instead return a new Struct with the FileInfo embedded in it.
	//
	// This will provide additional functionality to the FileInfo of
	// subsequent methods, as well as allow
	Ctx(string) interface{}
	SetCtx(string, interface{})
}

func NewFileInfo(p string) FileInfo {
	n := filepath.Base(p)
	d := filepath.Dir(p)
	return &fileInfo{
		name:         n,
		path:         d,
		originalName: n,
		originalPath: d,
		ctx:          make(map[string]interface{}),
	}
}

type fileInfo struct {
	name         string
	path         string
	originalName string
	originalPath string

	ctx map[string]interface{}
}

func (fi *fileInfo) Name() string {
	return fi.name
}

func (fi *fileInfo) SetName(s string) {
	fi.name = s
}

func (fi *fileInfo) Path() string {
	return fi.path
}

func (fi *fileInfo) SetPath(s string) {
	fi.path = s
}

func (fi *fileInfo) OriginalName() string {
	return fi.originalName
}

func (fi *fileInfo) OriginalPath() string {
	return fi.originalPath
}

func (fi *fileInfo) Ctx(k string) interface{} {
	return fi.ctx[k]
}

func (fi *fileInfo) SetCtx(k string, v interface{}) {
	fi.ctx[k] = v
}

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
