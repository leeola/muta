package muta

import "io"

const funcPluginName string = "muta.FuncStreamer"

// A FuncStreamer is a single Function implementation of a Streamer. Best
// used only for very simplistic Streamers that do not need to store any
// additional state values.
//
// Like a normal Streamer, each time `Next()` is called, this function
// is called. It is also passed the Streamer of the instance it represents.
// This is to be used to check the value of it's parent Streamer,
// `s.Streamer` as well as call embedded methods, such as
// `s.FrontMatter()`.
//
// Everything you would do in a normal Next() method, you do here as well.
// This just saves you from having to create a full `Use()`-able struct.
type FuncStreamer func(Streamer) (*FileInfo, io.ReadCloser, error)

// Create a new `Use()`-able FuncStream. See FuncStreamer docs for
// further explanation.
func FuncStream(fn FuncStreamer) StreamEmbedder {
	return StreamEmbedderFunc(func(inner Streamer) Streamer {
		return &funcStream{
			Streamer:     inner,
			funcStreamer: fn,
		}
	})
}

type funcStream struct {
	Streamer
	funcStreamer FuncStreamer
}

func (s *funcStream) Next() (*FileInfo, io.ReadCloser, error) {
	return s.funcStreamer(s)
}

func (s *funcStream) Use(embedder StreamEmbedder) Streamer {
	return embedder.Embed(s)
}
