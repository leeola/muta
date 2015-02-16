package muta

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

func NewFileInfo(p string) *FileInfo {
	n := filepath.Base(p)
	d := filepath.Dir(p)
	return &FileInfo{
		Name:         n,
		Path:         d,
		OriginalName: n,
		OriginalPath: d,
		Ctx:          make(map[string]interface{}),
	}
}

type FileInfo struct {
	Name         string
	Path         string
	OriginalName string
	OriginalPath string
	Ctx          map[string]interface{}
}

type SrcOpts struct {
	ReadSize uint
}

type Streamer func(*FileInfo, []byte) (*FileInfo, []byte, error)

// A convenience function to let functions that return Streamers
// "return an error". Ie, the following syntax:
//
// ```golang
// err := doSomething()
// if err != nil {
//   return ErrorStreamer(err)
// }
// ```
//
// ErrorStreamer will simply return a Streamer that will return an
// error when called.
func ErrorStreamer(err error) Streamer {
	return func(fi *FileInfo, chunk []byte) (*FileInfo, []byte, error) {
		return fi, chunk, err
	}
}

func SrcStreamer(ps []string, opts SrcOpts) Streamer {
	if opts.ReadSize == 0 {
		opts.ReadSize = 50
	}

	// Setup our channels
	fi := make(chan *FileInfo)
	chunk := make(chan []byte)
	err := make(chan error)
	read := make(chan bool)

	// This method of reading files needs to be abstracted further
	// to ensure that the file closing is deferred. In this
	// implementation i can't think of a way to test that.
	// Also, moving it out would let us ensure closing of the files
	// in tests
	go func() {
		sendErr := func(_fi *FileInfo, _chunk []byte, _err error) {
			<-read
			fi <- _fi
			chunk <- _chunk
			err <- _err
		}

		loadFile := func(p string) error {
			pchunks := make([]byte, opts.ReadSize)
			pfi := NewFileInfo(p)

			f, ferr := os.Open(p)
			defer f.Close()
			if ferr != nil {
				sendErr(pfi, nil, ferr)
				return ferr
			}

			// Wait for a read request
			for <-read {
				// Read
				count, ferr := f.Read(pchunks)
				if ferr != nil && ferr == io.EOF {
					break
				}

				// Send
				fi <- pfi
				chunk <- pchunks[0:count]
				err <- ferr
				if ferr != nil {
					return ferr
				}
			}

			// The for loop stopped, send EOF
			fi <- pfi
			chunk <- nil
			err <- nil
			return nil
		}

		// Go through the paths and globbify any globbed paths
		globbedPaths := []string{}
		for _, p := range ps {
			// If it hs a *, it is a glob
			if strings.Contains(p, "*") {
				expandedGlobs, err := filepath.Glob(p)
				if err != nil {
					sendErr(nil, nil, err)
					return
				}
				globbedPaths = append(globbedPaths, expandedGlobs...)
			} else {
				globbedPaths = append(globbedPaths, p)
			}
		}

		for _, p := range globbedPaths {
			err := loadFile(p)
			if err != nil {
				return
			}
		}

		<-read
		// send EOS
		fi <- nil
		chunk <- nil
		err <- nil
	}()

	return func(inFi *FileInfo, inC []byte) (*FileInfo, []byte, error) {
		// If there is an incoming file pass the data along unmodified. This
		// func doesn't care to modify the data in any way
		if inFi != nil {
			return inFi, inC, nil
		}

		read <- true
		return <-fi, <-chunk, <-err
	}
}
