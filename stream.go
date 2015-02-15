package muta

import "errors"

type Stream struct {
	Streamers []Streamer
}

func (s *Stream) Pipe(f Streamer) *Stream {
	s.Streamers = append(s.Streamers, f)
	return s
}

func (s *Stream) Start() {
	for i, fn := range s.Streamers {
		// In the near future, we need to check for errors returned by
		// the Streamer. Ignoring them for now.
		s.startGenerator(fn, s.Streamers[i+1:])
	}
}

// TODO: Split this function into smaller logical sections, for ease
// of debugging
func (s *Stream) startGenerator(generator Streamer,
	receivers []Streamer) (err error) {
	for true {
		genFi, genChunk, err := generator(nil, nil)
		if err != nil {
			return err
		}

		// Generator signalled EOS
		if genFi == nil {
			return nil
		}

		var fi *FileInfo
		var chunk []byte
		var streamed bool

		var fn Streamer
		for i := 0; i < len(receivers); i++ {
			if i == 0 {
				fi = genFi
				chunk = genChunk
				streamed = false
			}

			fn = receivers[i]
			fi, chunk, err = fn(fi, chunk)
			if err != nil {
				return err
			}

			// If any receiver streamed data, set streamed = true
			if chunk != nil {
				streamed = true
			}

			// Receiver signaled EOS
			if fi == nil {
				if streamed && genChunk == nil {
					// If any Receiver returned any data, we need to repeat
					// the stream until they all (up until EOS atleast) return
					// EOF
					// Repeat the stream
					i = -1
					continue
				} else if chunk == nil || genChunk != nil {
					// chunk == nil:
					// Receiver signaled EOS and EOF
					//
					// genChunk != nil
					// Receiver signaled EOS but not EOF, **and** the Generator
					// is not signaling EOF (ie, it's still streaming) data.
					// Repeating here would be bad, since it would repeat
					// the generator data.
					//
					// Stop the stream
					break
				} else {
					// Receiver signaled EOS and but not EOF
					// Repeat the stream
					i = -1
					continue
				}
			}

			// For now, if the receiver returns a different file than the
			// generator, error out. We're not yet supporting receiver->gen
			// converting.
			if fi != genFi {
				return errors.New("Receivers turning into Generators is not " +
					"yet implemented")
			}
		}
	}
	return nil
}

func Src(srcs ...string) *Stream {
	s := &Stream{}
	return s.Pipe(SrcStreamer(srcs, SrcOpts{}))
}
