package muta

import (
	"io"
	"os"
	"path/filepath"
)

type FileInfo struct {
	Name string
	Path string
	Ctx  *interface{}
}

type SrcOpts struct {
	ReadSize uint
}

type Streamer func(*FileInfo, []byte) (*FileInfo, []byte, error)

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
		loadFile := func(p string) {
			pchunks := make([]byte, opts.ReadSize)
			pfi := &FileInfo{
				Name: filepath.Base(p),
				Path: filepath.Dir(p),
			}

			f, ferr := os.Open(p)
			defer f.Close()
			if ferr != nil {
				fi <- pfi
				chunk <- nil
				err <- ferr
				return
			}

			for true {
				// Wait for a read request
				<-read

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
					return
				}
			}

			fi <- pfi
			chunk <- nil
			err <- nil
		}

		for _, p := range ps {
			loadFile(p)
		}
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

func Dest(d string) Streamer {
	return func(fi *FileInfo, chunk []byte) (*FileInfo, []byte, error) {
		return nil, nil, nil
	}
}
