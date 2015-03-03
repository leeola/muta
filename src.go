package muta

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/leeola/muta/logging"
)

func Src(srcs ...string) *Stream {
	s := &Stream{}
	return s.Pipe(NewSrcStreamer(srcs...))
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

func NewSrcStreamer(srcs ...string) *SrcStreamer {
	return (&SrcStreamer{Sources: srcs}).Init()
}

type SrcStreamer struct {
	// The chunk size that this Streamer will return, as it's reading
	// a file
	ReadSize uint

	// The base directory that will be trimmed from the output path.
	// For example, `SrcStreamer("foo/bar/baz")` would set a Base of
	// `"foo/bar"`, so that the FileInfo has a Path of `.`. Trimming
	// `"foo/bar"` from the Path. You can override this, by setting
	// this value manually.
	Base string

	// The filepaths that this Streamer will load, and Stream
	Sources []string

	// The name of the Streamer (returned by the s.Name() interface)
	name string

	fi *FileInfo
	f  *os.File

	// The current working index of the Sources slice
	sourceIndex int
}

func (s *SrcStreamer) Init() *SrcStreamer {
	// Clean the paths to remove any oddities (before setting opts)
	for i, p := range s.Sources {
		s.Sources[i] = filepath.Clean(p)
	}

	// Set default options
	if s.name == "" {
		s.name = "muta.Src"
	}
	if s.ReadSize == 0 {
		s.ReadSize = 50
	}
	if s.Base == "" {
		s.Base = globsToBase(s.Sources...)
	}

	return s
}

func (s *SrcStreamer) Name() string {
	return s.name
}

func (s *SrcStreamer) openNext() (*os.File, *FileInfo, error) {
	if len(s.Sources) == 0 {
		return nil, nil, nil
	}

	// If the next source has a glob in it, expand the glob
	if strings.Contains(s.Sources[0], "*") {
		expanded, err := filepath.Glob(s.Sources[0])
		if err != nil {
			return nil, nil, err
		}
		s.Sources = append(expanded, s.Sources[1:]...)
	}
	p := s.Sources[0]
	s.Sources = s.Sources[1:]

	// Create the fileinfo, and trim the base from the destination path
	fi := NewFileInfo(p)
	fi.Path = strings.TrimPrefix(fi.Path, s.Base)
	if fi.Path == "" {
		fi.Path = "."
	}

	logging.Debug([]string{s.name}, "Opening", p)
	// Open the file for reading
	f, err := os.Open(p)
	if err != nil {
		return nil, fi, err
	}

	return f, fi, nil
}

// Read bytes from the current *os.File
func (s *SrcStreamer) read() (*FileInfo, []byte, error) {
	// If s.f is nil, open the next file
	if s.f == nil {
		f, fi, err := s.openNext()
		if f == nil || err != nil {
			return nil, nil, err
		}
		s.f = f
		s.fi = fi
	}

	chunk := make([]byte, s.ReadSize)
	count, err := s.f.Read(chunk)

	switch {
	case err == io.EOF:
		s.f.Close()
		s.f = nil
		return s.fi, nil, nil

	case err != nil:
		return s.fi, nil, err

	default:
		return s.fi, chunk[:count], nil
	}
}

func (s *SrcStreamer) Stream(fi *FileInfo, chunk []byte) (*FileInfo, []byte, error) {
	// We're not modifying incoming data at all. Pass through everything
	if fi != nil {
		return fi, chunk, nil
	}

	return s.read()
}
