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

func (s *Stream) startGenerator(generator Streamer,
	receivers []Streamer) (err error) {
	for true {
		genFi, chunk, err := generator(nil, nil)
		if err != nil {
			return err
		}

		// Generator signalled EOS
		if genFi == nil {
			return nil
		}

		fi := genFi
		for _, fn := range receivers {
			fi, chunk, err = fn(fi, chunk)
			if err != nil {
				return err
			}

			// A receiver is signaling EOS, so stop the fi/chunk propagation
			if fi == nil {
				break
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
