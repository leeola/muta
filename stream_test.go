package muta

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestStreamPipe(t *testing.T) {
	Convey("Should add the Streamer to the Streamers", t, func() {
		s := Stream{}
		sr := func(fi *FileInfo, chunk []byte) (*FileInfo, []byte, error) {
			return fi, chunk, nil
		}
		s.Pipe(sr)
		So(s.Streamers, ShouldContain, sr)
	})
}

func TestStreamstartGenerator(t *testing.T) {
	Convey("Should call the Generator with EOS to signal that it is "+
		"the next Generator", t, func() {
		calls := []string{}
		var genFi *FileInfo
		var genChunk []byte
		a := func(fi *FileInfo, chunk []byte) (*FileInfo, []byte, error) {
			calls = append(calls, "a")
			genFi = fi
			genChunk = chunk
			return nil, nil, nil
		}
		s := Stream{}
		s.startGenerator(a, []Streamer{})
		So(calls, ShouldResemble, []string{"a"})
		So(genFi, ShouldBeNil)
		So(genChunk, ShouldBeNil)
	})

	Convey("Should repeatedly call the Generator", t, func() {
		Convey("until it signals EOS", func() {
			calls := []string{}
			a := func(_ *FileInfo, _ []byte) (*FileInfo, []byte, error) {
				calls = append(calls, "a")
				if len(calls) < 3 {
					return &FileInfo{}, nil, nil
				} else {
					return nil, nil, nil
				}
			}
			s := Stream{}
			s.startGenerator(a, []Streamer{})
			So(calls, ShouldResemble, []string{"a", "a", "a"})
		})

		Convey("unless it returns an error", func() {
			calls := []string{}
			genErr := errors.New("test error")
			a := func(_ *FileInfo, _ []byte) (*FileInfo, []byte, error) {
				calls = append(calls, "a")
				return &FileInfo{}, []byte("foo"), genErr
			}
			s := Stream{}
			err := s.startGenerator(a, []Streamer{})
			So(calls, ShouldResemble, []string{"a"})
			So(err, ShouldEqual, genErr)
		})
	})

	Convey("Should stop a Stream when a receiver signals EOS", t, func() {
		calls := []string{}
		genFi := &FileInfo{}
		a := func(_ *FileInfo, _ []byte) (*FileInfo, []byte, error) {
			calls = append(calls, "a")
			// Return a valid fi the first time, signal EOS the second+ time
			fi := genFi
			genFi = nil
			return fi, nil, nil
		}
		b := func(_ *FileInfo, _ []byte) (*FileInfo, []byte, error) {
			calls = append(calls, "b")
			return nil, nil, nil
		}
		c := func(_ *FileInfo, _ []byte) (*FileInfo, []byte, error) {
			calls = append(calls, "c")
			return nil, nil, nil
		}
		s := Stream{}
		s.startGenerator(a, []Streamer{b, c})
		So(calls, ShouldResemble, []string{"a", "b", "a"})
	})

	Convey("Should pass EOF to all Streamers", t, func() {
		calls := []string{}
		chunks := [][]byte{}
		a := func(_ *FileInfo, _ []byte) (*FileInfo, []byte, error) {
			calls = append(calls, "a")
			if len(calls) == 1 {
				fi := &FileInfo{
					Name: "foo",
				}
				return fi, nil, nil
			} else {
				return nil, nil, nil
			}
		}
		b := func(fi *FileInfo, chunk []byte) (*FileInfo, []byte, error) {
			calls = append(calls, "b")
			chunks = append(chunks, chunk)
			return fi, chunk, nil
		}
		c := func(fi *FileInfo, chunk []byte) (*FileInfo, []byte, error) {
			calls = append(calls, "c")
			chunks = append(chunks, chunk)
			return fi, chunk, nil
		}
		s := Stream{}
		s.startGenerator(a, []Streamer{b, c})
		So(calls, ShouldResemble, []string{"a", "b", "c", "a"})
		So(chunks, ShouldResemble, [][]byte{nil, nil})
	})
}

func TestStreamStart(t *testing.T) {
}

func TestSrc(t *testing.T) {
}
