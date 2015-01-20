package muta

type Stream struct {
	Streamers []Streamer
}

func (s *Stream) Pipe(f Streamer) *Stream {
	s.Streamers = append(s.Streamers, f)
	return s
}

func (s *Stream) pipeChunk(streamers []Streamer,
	fi *FileInfo, chunk []byte) (err error) {
	for _, fn := range streamers {
		fi, chunk, err = fn(fi, chunk)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Stream) Start() {
	for i, fn := range s.Streamers {
		fi, chunk, _ := fn(nil, nil)
		// In the near future, we need to check for errors returned by
		// the Streamer. Ignoring them for now.
		s.pipeChunk(s.Streamers[i+1:], fi, chunk)
	}
}

func Src(srcs ...string) *Stream {
	s := &Stream{}
	return s.Pipe(SrcStreamer(srcs, SrcOpts{}))
}
