package muta

import "io"

// NewStream simply returns a Streamer slice, as a Stream type. This
// simply exists for convention.
func NewStream() Stream {
	return []Streamer{}
}

// A Stream is simply a slice of Streamers, which will
// chain Next calls from the beginning of the Stream to the end.
//
// It's worth noting that the Stream is infact a Streamer itself, as
// it satisfies the Streamer interface. As such, a Stream can be piped to
// or from another Streamer and/or Stream. Turtles all the way down.
type Stream []Streamer

// Pipe appends the given Streamer to the slice, then returning the
// slice.
//
// If the Streamer implements error the slice is resized to only contain
// the error Streamer. Meaning that the error Streamer will be called
// first when this Stream is started, and in theory, also returning an
// error on the first Next() call.
func (s Stream) Pipe(sr Streamer) Stream {
	if _, ok := sr.(error); ok {
		s[0] = sr
		return s[:1]
	}

	return append(s, sr)
}

// Next satisifies the Streamer interface by providing any incoming
// FileInfo and ReadCoser to all of the Streamer's contained in this
// Stream.
func (s Stream) Next(inFi FileInfo, inRc io.ReadCloser) (fi FileInfo,
	rc io.ReadCloser, err error) {

	fi = inFi
	rc = inRc
	for _, sr := range s {
		fi, rc, err = sr.Next(fi, rc)
		if err != nil {
			return
		}
	}
	return
}

// Stream calls all of the Streamer's until they return no files
func (s Stream) Stream() error {
	for {
		fi, rc, err := s.Next(nil, nil)

		if rc != nil {
			rc.Close()
		}

		if err != nil {
			return err
		}

		if fi == nil {
			return nil
		}
	}
}
