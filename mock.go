//
// # Mock
//
// A series of testing aids.
//
package muta

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/leeola/muta/mutil"
)

// A streamer that creates files and contents, based on the Files
// and Contents slices.
//
// TODO: Find a way to move this into the `muta/mtesting` package. The
// problem is that if this is in `muta/mtesting`, then the signature of
// MockStreamer.Next becomes `Next() (*muta.FileInfo ...)`. SrcStreamer
// and DestStreamer however, require that the signature of any locally
// embedded library is `Next() (*FileInfo ...)` instead.
//
// When `muta` is imported into another library (such as external
// Streamers), this appears to not be an issue, because references to
// SrcStreamer become `muta.SrcStreamer`, and signatures become
// `muta.FileInfo`, and so on.
//
// I could be way off base though - i'm not sure what to do here.
type MockStreamer struct {
	Streamer

	// A slice of the file names to generate. If no Content is provided,
	// for an individual file (eg, if there are 5 files, but 4 contents)
	// the content will be automatically created as `<filename> content`
	Files []string

	// An optional slice of the contents to use for each file. These are
	// Shifted off one by one in the same order as files. If, after the slice
	// is empty, there are still more files, the contents are automatically
	// created as `<filename> content`.
	//
	// Additionally, if the contents of a file begin with `error: ` then
	// an error is returned for that file, with any remaining characters
	// of the content string. This allows you to mock errors, from a
	// Streamer as well.
	Contents []string
}

func (s *MockStreamer) Use(embedder StreamEmbedder) Streamer {
	return embedder.Embed(s)
}

func (s *MockStreamer) Next() (fi *FileInfo, rc io.ReadCloser, err error) {
	if s.Streamer != nil {
		fi, rc, err = s.Streamer.Next()
		if fi != nil {
			return
		}
	}

	if len(s.Files) == 0 {
		return
	}

	file := s.Files[0]
	s.Files = s.Files[1:]

	// Shift content off of the list if there is any, or create content
	// if there isn't any. Note that we don't check for an empty string,
	// which lets you pass an "empty" file with `Contents: []string{""}`
	var content string
	if len(s.Contents) > 0 {
		content = s.Contents[0]
		s.Contents = s.Contents[1:]
	} else {
		content = fmt.Sprintf("%s content", file)
	}

	fi = NewFileInfo(file)
	rc = mutil.ByteCloser([]byte(content))

	// If the mock content starts with `error: `, return an error
	// for this file.
	if strings.HasPrefix(content, "error: ") {
		err = errors.New(fmt.Sprintf(
			"MockStreamer Mock Error: %s",
			strings.TrimLeft(content, "error: ")))
	}
	return
}
