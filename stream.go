package muta

type Stream struct {
	Streamers []Streamer
}

func (s *Stream) Pipe(f Streamer) *Stream {
	s.Streamers = append(s.Streamers, f)
	return s
}

func (s *Stream) Start() {
	for _, f := range s.Streamers {
		// Call the first Streamer with nil values. This signals it to
		// start generating it's own files (if any).
		f(nil, nil)
		// In the near future, we need to check for errors returned by
		// the Streamer. Ignoring them for now.
	}
}

func Src(srcs ...string) *Stream {
	s := &Stream{}
	return s.Pipe(SrcStreamer(srcs...))
}
