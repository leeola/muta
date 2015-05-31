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
// For an use-case understanding of this func, see the SrcStream.Base
// docstring.
//
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

// Return a new SrcStream instance. If you need a Use() able version of
// Src, see UsableSrc()
func Src(paths ...string) Streamer {
	return (&SrcStream{Sources: paths}).init()
}

func UsableSrc(paths ...string) StreamEmbedder {
	return StreamEmbedderFunc(func(inner Streamer) Streamer {
		s := &SrcStream{
			Streamer: inner,
			Sources:  paths,
		}
		return s.init()
	})
}

type SrcStream struct {
	// The optional inner Streamer.
	Streamer

	// The base directory that will be trimmed from the output path.
	// For example, `SrcStream("foo/bar/baz")` would set a Base of
	// `"foo/bar"`, so that the FileInfo has a Path of `.`. Trimming
	// `"foo/bar"` from the Path. You can override this, by setting
	// this value manually.
	Base string

	// The filepaths that this Streamer will load, and Stream
	Sources []string
}

func (s *SrcStream) init() *SrcStream {
	// Clean the paths to remove any oddities (before setting opts)
	for i, p := range s.Sources {
		s.Sources[i] = filepath.Clean(p)
	}

	if s.Base == "" {
		s.Base = globsToBase(s.Sources...)
	}

	return s
}

func (s *SrcStream) Use(embedder StreamEmbedder) Streamer {
	return embedder.Embed(s)
}

func (s *SrcStream) Next() (*FileInfo, io.ReadCloser, error) {
	if s.Streamer != nil {
		if fi, r, err := s.Streamer.Next(); fi != nil || err != nil {
			return fi, r, nil
		}
	}

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

	fi := NewFileInfo(p)
	fi.Path = strings.TrimPrefix(fi.Path, s.Base)
	if fi.Path == "" {
		fi.Path = "."
	}

	logging.Debug([]string{srcPluginName}, "Opening", p)
	// Open the file for reading
	f, err := os.Open(p)

	if err != nil {
		return fi, f, err
	}

	return fi, f, nil
}
