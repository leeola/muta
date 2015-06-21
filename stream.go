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
func (s Stream) Next(fi FileInfo, rc io.ReadCloser) (FileInfo,
	io.ReadCloser, error) {

	return s.NextFrom(0, fi, rc)
}

// NextFrom takes the give FileInfo and io.ReadCloser and pipes it
// through this Streams Streamers, starting from the given index.
//
// If any Streamers return a nil file, no further Streamers are called.
//
// NextFrom is mostly an implementation detail, but is public to allow you
// to step through the slice at various points. Useful for debugging,
// testing, etc.
func (s Stream) NextFrom(from int, inFi FileInfo, inRc io.ReadCloser) (
	fi FileInfo, rc io.ReadCloser, err error) {

	fi = inFi
	rc = inRc
	for ; from < len(s); from++ {
		fi, rc, err = s[from].Next(fi, rc)

		if err != nil {
			return
		}

		if fi == nil {
			return
		}
	}
	return
}

// Stream calls all of the Streamer's until every Streamer has stopped
// returning files.
//
// Each Streamer is called with `nil,nil`. If the called Streamer returns
// a FileInfo, the return values are passed onto the next Streamers, and
// the Streamer will be called again. This will repeat, until the Streamer
// returns a nil FileInfo. Once that happens, the next Streamer in the
// slice is treated the same way.
func (s Stream) Stream() (err error) {
	var fi FileInfo
	var rc io.ReadCloser

	for i := 0; i < len(s); i++ {
		// Call the current Streamer
		fi, rc, err = s[i].Next(nil, nil)

		if err != nil {
			return err
		}

		// If the current Streamer returned a nil FileInfo, move onto the
		// next Streamer.
		if fi == nil {
			continue
		}

		// Pass the Streamers return values onto all the other Streamers.
		// Note that we're using the index+1, to ensure the Current Streamer
		// isn't passed it's own returned file.
		_, rc, err = s.NextFrom(i+1, fi, rc)

		// Since the current Streamer returned a FileInfo, move the index back
		// so that it is called again, and again, until it finally returns
		// no more files.
		i--

		// If an ReadCloser was returned, Close it to be safe.
		if rc != nil {
			rc.Close()
		}

		// The other Streamers returned an error
		if err != nil {
			return err
		}
	}

	return
}
