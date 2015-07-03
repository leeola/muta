package muta

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/leeola/muta/logging"
)

const srcPluginName string = "muta.Src"

// globsToBase, will take a series of globs and find the base among
// all of the given globs.
//
// For a use-case understanding of this func, see the SrcStreamer.Base
// docstring.
//
// Note that we're just supporting the * glob star currently, as i believe
// that's the only supported glob pattern, from the official lib.
// Return a new Stream, with a SrcStreamer. If you need a Pipe()able
// version of Src, see PipeableSrc()
func Src(paths ...string) Stream {
	return []Streamer{PipeableSrc(paths...)}
}

// PipeableSrc
func PipeableSrc(paths ...string) *SrcStreamer {
	return (&SrcStreamer{Sources: paths}).init()
}

type SrcStreamer struct {
	// The base directory that will be trimmed from the output path.
	// For example, `SrcStreamer("foo/bar/baz")` would set a Base of
	// `"foo/bar"`, so that the FileInfo has a Path of `.`. Trimming
	// `"foo/bar"` from the Path. You can override this, by setting
	// this value manually.
	Base string

	// The filepaths that this Streamer will load, and Stream
	Sources []string
}

func (s *SrcStreamer) init() *SrcStreamer {
	// Clean the paths to remove any oddities (before setting opts)
	for i, p := range s.Sources {
		s.Sources[i] = filepath.Clean(p)
	}

	if s.Base == "" {
		s.Base = globsToBase(s.Sources...)
	}

	return s
}

func (s *SrcStreamer) Next(fi FileInfo, rc io.ReadCloser) (FileInfo,
	io.ReadCloser, error) {

	// If file's are incoming, return them. Src does not need to
	// modify them.
	if fi != nil {
		return fi, rc, nil
	}

	// If there are no source files to generate, return nil.
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

	// Shift a path from the Sources slice
	p := s.Sources[0]
	s.Sources = s.Sources[1:]

	fi = NewFileInfo(p)
	fi.SetPath(strings.TrimPrefix(fi.Path(), s.Base))
	if fi.Path() == "" {
		fi.SetPath(".")
	}

	logging.Debug([]string{srcPluginName}, "Opening", p)
	// Open the file for reading
	f, err := os.Open(p)

	if err != nil {
		return fi, f, err
	}

	return fi, f, nil
}

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
