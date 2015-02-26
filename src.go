package muta

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/leeola/muta/logging"
)

type SrcOpts struct {
	Name     string
	ReadSize uint

	// The base directory that will be trimmed from the output path.
	// For example, `SrcStreamer("foo/bar/baz")` would set a Base of
	// `"foo/bar"`, so that the FileInfo has a Path of `.`. Trimming
	// `"foo/bar"` from the Path. You can override this, by setting
	// this value manually.
	Base string
}

func Src(srcs ...string) *Stream {
	s := &Stream{}
	return s.Pipe(SrcStreamer(srcs, SrcOpts{}))
}

// this func will be moved into it's own Package, and probably it's own
// repo in the future
// Note that we're just supporting the * glob star currently, as i believe
// that's the only supported glob pattern, from the official lib.
func globsToBase(globs ...string) string {
	var base string
	var depth int
	var globChars string = "*"

	for _, glob := range globs {
		var gBase string
		var gDepth int
		if i := strings.IndexAny(glob, globChars); i > -1 {
			gBase = filepath.Dir(glob[:i])
		} else {
			gBase = filepath.Dir(glob)
		}
		gDepth = strings.Count(gBase, string(filepath.Separator))
		// If there is no base, or
		// if the glob's depth is smaller (more base) than the depth
		if base == "" || gDepth < depth {
			base = gBase
			depth = gDepth
		}
	}
	return base
}

func SrcStreamer(ps []string, opts SrcOpts) Streamer {
	// Clean the paths to remove any oddities (before setting opts)
	for i, p := range ps {
		ps[i] = filepath.Clean(p)
	}

	// Set default options
	if opts.Name == "" {
		opts.Name = "muta.Src"
	}
	if opts.ReadSize == 0 {
		opts.ReadSize = 50
	}
	if opts.Base == "" {
		opts.Base = globsToBase(ps...)
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
			// Trim the base from the file
			pfi.Path = strings.TrimPrefix(pfi.Path, opts.Base)
			if pfi.Path == "" {
				pfi.Path = "."
			}

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
			logging.Debug([]string{opts.Name}, "Reading", p)
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

	return NewEasyStreamer(opts.Name, func(inFi *FileInfo,
		inC []byte) (*FileInfo, []byte, error) {
		// If there is an incoming file pass the data along unmodified. This
		// func doesn't care to modify the data in any way
		if inFi != nil {
			return inFi, inC, nil
		}

		read <- true
		return <-fi, <-chunk, <-err
	})
}
