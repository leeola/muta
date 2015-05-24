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

	"github.com/leeola/muta/mutil"
)

// A streamer that creates files and contents, based on the input name
type MockStreamer struct {
	Streamer
	Mocks []string
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

	if len(s.Mocks) == 0 {
		return
	}

	mock := s.Mocks[0]
	s.Mocks = s.Mocks[1:]

	fi = NewFileInfo(mock)
	rc = mutil.ByteCloser([]byte(fmt.Sprintf("%s content", mock)))
	// If the mock has the name error, return an error
	if mock == "error" {
		err = errors.New(fmt.Sprintf("MockStreamer Mock Error"))
	}
	return
}
