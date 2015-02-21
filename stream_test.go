package muta

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// A little utility func
func ContainsStringCount(sl []string, s string) (c int) {
	for _, item := range sl {
		if item == s {
			c++
		}
	}
	return c
}

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
		err := s.startGenerator(a, []Streamer{b, c})
		So(err, ShouldBeNil)
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

	// Might be removing this "feature". I'm not positive that it's
	// existence justifies the complexity added to the stream API
	//Convey("When a Streamer signals EOS but not EOF,", t, func() {
	//	Convey("Repeat the stream until they signal EOF", func() {
	//		calls := []string{}
	//		originalFi := &FileInfo{Name: "foo"}
	//		gen := func(_ *FileInfo, _ []byte) (*FileInfo, []byte, error) {
	//			calls = append(calls, "gen")
	//			switch ContainsStringCount(calls, "gen") {
	//			case 1:
	//				return originalFi, nil, nil
	//			default:
	//				return nil, nil, nil
	//			}
	//		}
	//		a := func(fi *FileInfo, chunk []byte) (*FileInfo, []byte, error) {
	//			calls = append(calls, "a")
	//			if len(calls) <= 3 {
	//				// By returning EOS but not EOF, we cause the stream to
	//				// repeat. It stops here for now, but it repeats the last
	//				// generator return until this func returns a nil byte
	//				return nil, []byte{}, nil
	//			} else {
	//				return fi, chunk, nil
	//			}
	//		}
	//		b := func(fi *FileInfo, chunk []byte) (*FileInfo, []byte, error) {
	//			calls = append(calls, "b")
	//			return fi, chunk, nil
	//		}
	//		s := Stream{}
	//		s.startGenerator(gen, []Streamer{a, b})
	//		So(calls, ShouldResemble, []string{
	//			"gen", "a", "a", "a", "b", "gen"})
	//	})

	//	Convey("Do not repeat the Stream if the Generator did not "+
	//		"return EOF", func() {
	//		calls := []string{}
	//		originalFi := &FileInfo{Name: "foo"}
	//		gen := func(_ *FileInfo, _ []byte) (*FileInfo, []byte, error) {
	//			calls = append(calls, "gen")
	//			switch ContainsStringCount(calls, "gen") {
	//			case 1:
	//				return originalFi, []byte("foo"), nil
	//			case 2:
	//				return originalFi, nil, nil
	//			default:
	//				return nil, nil, nil
	//			}
	//		}
	//		a := func(fi *FileInfo, chunk []byte) (*FileInfo, []byte, error) {
	//			calls = append(calls, "a")
	//			if ContainsStringCount(calls, "a") == 1 {
	//				// This normally signals stream repeat, but it should not
	//				// repeat here because the Generator did not signal EOS
	//				// EOF
	//				return nil, []byte{}, nil
	//			} else {
	//				return fi, chunk, nil
	//			}
	//		}
	//		b := func(fi *FileInfo, chunk []byte) (*FileInfo, []byte, error) {
	//			calls = append(calls, "b")
	//			return fi, chunk, nil
	//		}
	//		s := Stream{}
	//		s.startGenerator(gen, []Streamer{a, b})
	//		So(calls, ShouldResemble, []string{
	//			"gen", "a", // "a" signals EOS (with a repeat []byte)
	//			"gen", "a", "b", "gen"})
	//	})
	//})

	Convey("When a Generator signals EOS and EOF", t, func() {
		Convey("Repeat the Stream if any Receiver returns bytes", func() {
			calls := []string{}
			data := []byte{}
			originalFi := &FileInfo{Name: "foo"}
			gen := func(_ *FileInfo, _ []byte) (*FileInfo, []byte, error) {
				calls = append(calls, "gen")
				if ContainsStringCount(calls, "gen") == 1 {
					return originalFi, nil, nil
				} else {
					return nil, nil, nil
				}
			}
			a := func(fi *FileInfo, _ []byte) (*FileInfo, []byte, error) {
				calls = append(calls, "a")
				return fi, nil, nil
			}
			b := func(fi *FileInfo, _ []byte) (*FileInfo, []byte, error) {
				calls = append(calls, "b")
				if ContainsStringCount(calls, "b") <= 2 {
					return fi, []byte("foo"), nil
				} else {
					return fi, nil, nil
				}
			}
			c := func(fi *FileInfo, chunk []byte) (*FileInfo, []byte, error) {
				calls = append(calls, "c")
				data = append(data, chunk...)
				return nil, nil, nil
			}
			s := Stream{}
			err := s.startGenerator(gen, []Streamer{a, b, c})
			So(err, ShouldBeNil)
			So(calls, ShouldResemble, []string{
				"gen", "a", "b", "c", // b returned data
				"a", "b", "c", // b returned data
				"a", "b", "c", // b returned EOF
				"gen", // gen returned EOS
			})
			So(string(data), ShouldResemble, "foofoo")
		})
	})
}

func TestStreamStart(t *testing.T) {
}

func TestSrc(t *testing.T) {
}
