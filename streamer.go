package muta

import (
	"io"
	"path/filepath"
	"strings"
)

type Streamer interface {
	Next() (*FileInfo, io.ReadCloser, error)
	Use(StreamEmbedder) Streamer
}

// The StreamEmbedder interface enables one Streamer to be created
// while embedding the previous Stream, inside itself. There is no
// contract that this will happen, the StreamEmbedder simply enables the
// functionality.
//
// NOTE: A StreamEmbedder cannot return an Error (only a Streamer). For
// an explanation of why _(and how to return errors)_, see the
// ErrorStreamer docstring.
//
// To further illustrate why a StreamEmbedder is needed, lets look at the
// intended UX of the Streamer interface:
//
// ```
// Streamer()
// 	.Use(OtherStreamer())
// 	.Use(FinalStreamer())
// ```
//
// Remember that each Streamer is embedding the functionality of the
// previous Streamer. Enabling a plugin to cascade it's functionality,
// based on the data Stream.
//
// Without the StreamEmbedder, we would have to store references of each
// Streamer to pass it into the creation of the next streamer, like:
//
// ```
// s := Streamer()
// o := OtherStreamer(s)
// f := finalStreamer(f)
// ```
//
// or invert the flow, to something like:
//
// ```
// FinalStreamer(
//	OtherStreamer(
//		Streamer()
//	)
// )
// ```
//
// These options degrate the experience of the actual Stream user's (ie,
// the people using the plugins). So, a little complexity in regards to
// `.Use()` and StreamEmbedder seems better for the final product.
type StreamEmbedder interface {
	Embed(Streamer) Streamer
}

// A function type that satisfies the StreamEmbedder interface, allowing
// a function to be passed in which will also satisfy the interface.
//
// This is identical to how http.Handler and http.HandlerFunc operate.
type StreamEmbedderFunc func(Streamer) Streamer

func (f StreamEmbedderFunc) Embed(s Streamer) Streamer {
	return f(s)
}

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

	Ctx map[string]interface{}
}

func NewErrorStreamer(msgs ...string) ErrorStreamer {
	return ErrorStreamer{strings.Join(msgs, " ")}
}

// The ErrorStreamer fulfills both the Error and Streamer interfaces.
// No actual Streamer functionality is built in, and any usage of this
// Streamer returns itslf as an Error.
//
// This is useful because the responsibility for error checking and
// preventing further Streamer creation _(via `Streamer.Use()`)_ would
// be on each Streamer created. Meaning that everyone, everywhere, would
// have to "play ball" with a potentially sublte implementation.
//
// Rather than place this burdon on them, we can use ErrorStreamer. If
// you encounter an error during the StreamEmbedder execution, wrap your
// error with a ErrorStreamer, and return it instead. No further Streamers
// will be created in the `.Use()` chain.
type ErrorStreamer struct {
	Message string
}

func (s ErrorStreamer) Error() string {
	return s.Message
}

func (s ErrorStreamer) Use(embedder StreamEmbedder) Streamer {
	// Don't call embedder.Embed(), just return this Streamer. This will simply
	// loop for every `.Use()` call, while ensuring that no new Streamers
	// are instantiated.
	return s
}

func (s ErrorStreamer) Next() (*FileInfo, io.ReadCloser, error) {
	return nil, nil, s
}
