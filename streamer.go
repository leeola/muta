package muta

import "path/filepath"

func NewFileInfo(p string) *FileInfo {
	n := filepath.Base(p)
	d := filepath.Dir(p)
	return &FileInfo{
		Name:         n,
		Path:         d,
		OriginalName: n,
		OriginalPath: d,
		Ctx:          make(map[string]interface{}),
	}
}

type FileInfo struct {
	Name         string
	Path         string
	OriginalName string
	OriginalPath string
	Ctx          map[string]interface{}
}

// An alias for NewEasyStreamer. I'm undecided which func name i want to use, currently.
func NewStreamer(s string, f func(*FileInfo, []byte) (*FileInfo, []byte, error)) Streamer {
	return NewEasyStreamer(s, f)
}

func NewEasyStreamer(s string, f func(*FileInfo, []byte) (*FileInfo, []byte, error)) Streamer {
	return &EasyStreamer{name: s, stream: f}
}

// The EasyStreamer implements the Streamer interace by using the embedded
// `name` and `stream` value and func. This is purely a convenience method,
// so Streamer implementors don't have to implement the Streamer interface
// if they don't want to.
type EasyStreamer struct {
	name   string
	stream func(*FileInfo, []byte) (*FileInfo, []byte, error)
	Streamer
}

func (es *EasyStreamer) Name() string {
	return es.name
}

func (es *EasyStreamer) Stream(fi *FileInfo, chunk []byte) (*FileInfo, []byte, error) {
	return es.stream(fi, chunk)
}

type Streamer interface {
	Stream(*FileInfo, []byte) (*FileInfo, []byte, error)
	Name() string
}

// A convenience function to let functions that return Streamers
// "return an error". Ie, the following syntax:
//
// ```golang
// err := doSomething()
// if err != nil {
//   return ErrorStreamer(err)
// }
// ```
//
// ErrorStreamer will simply return a Streamer that will return an
// error when called.
func ErrorStreamer(err error) Streamer {
	return NewStreamer("muta.ErrorStreamer", func(fi *FileInfo, chunk []byte) (*FileInfo, []byte, error) {
		return fi, chunk, err
	})
}

func PassThroughStreamer() Streamer {
	return NewEasyStreamer("muta.PassThroughStreamer", func(fi *FileInfo, chunk []byte) (*FileInfo, []byte, error) {
		return fi, chunk, nil
	})
}
