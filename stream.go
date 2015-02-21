package muta

import (
	"errors"
	"fmt"
)

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

// unknown
func runStreamer(steamer Streamer, streamers []Streamer, fi *FileInfo, chunk []byte) error {
	return nil
}

// Run data through a Stream, stopping when any of them
// signals EOS
func (s *Stream) streamData(streamers []Streamer, fi *FileInfo, chunk []byte) error {
	var sFi *FileInfo
	var err error
	for _, streamer := range streamers {
		sFi, chunk, err = streamer(fi, chunk)
		switch {
		case err != nil:
			return err
		case sFi == nil:
			return nil
		case sFi != fi:
			return errors.New("Not Implemented. At this time, the same file" +
				"that was given must be returned")
		}
	}
	return nil
}

// Stream EOF through the Streamers. If any Streamer returns data,
// that data is run through like normal... TODO.. write this lol
func (s *Stream) streamEOF(streamers []Streamer, fi *FileInfo) error {
	var sFi *FileInfo
	var chunk []byte
	var err error
	for repeatEOF := true; repeatEOF; {
		repeatEOF = false
		for i := 0; i < len(streamers); i++ {
			sFi, chunk, err = streamers[i](fi, chunk)
			if err != nil {
				return err
			}
			// If a streamer signaled EOS, exit the inner loop to repeat the
			// stream, if needed.
			if sFi == nil {
				break
			}
			// If a streamer returned a different file than it was given
			// error out. This is not currently supported.
			if sFi != fi {
				return errors.New(fmt.Sprintf(
					"Not Implemented. At this time, the same file that was "+
						"given must be returned. Original file '%s', returned "+
						"file '%'", fi.Name, sFi.Name))
			}
			if chunk != nil {
				repeatEOF = true
			}
		}
	}
	return nil
}

func (s *Stream) startGenerator(generator Streamer, receivers []Streamer) error {
	var fi *FileInfo
	var chunk []byte
	var err error
	for true {
		fi, chunk, err = generator(nil, nil)
		if err != nil {
			return err
		}

		if fi == nil {
			return nil
		}

		if chunk != nil {
			err = s.streamData(receivers, fi, chunk)
		} else {
			err = s.streamEOF(receivers, fi)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
