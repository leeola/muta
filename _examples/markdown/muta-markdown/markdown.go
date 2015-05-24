package mutamarkdown

import (
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/leeola/muta"
	"github.com/leeola/muta/mutil"
	"github.com/russross/blackfriday"
)

func Markdown() muta.StreamEmbedder {
	return muta.StreamEmbedderFunc(func(inner muta.Streamer) muta.Streamer {
		return &MarkdownStreamer{inner}
	})
}

type MarkdownStreamer struct {
	muta.Streamer
}

func (s *MarkdownStreamer) Use(embedder muta.StreamEmbedder) muta.Streamer {
	return embedder.Embed(s)
}

func (s *MarkdownStreamer) Next() (*muta.FileInfo, io.ReadCloser, error) {
	// We don't generate files, so no need to ever do anything if we don't
	// have an inner Streamer.
	if s.Streamer == nil {
		return nil, nil, nil
	}

	fi, r, err := s.Streamer.Next()
	if fi == nil || err != nil {
		return fi, r, err
	}

	// If the file isn't markdown, we don't care about it. Return it
	// unmodified.
	if filepath.Ext(fi.Name) != ".md" {
		return fi, r, err
	}

	// Since the file is markdown, read it all so we can convert it to
	// markdown.
	markdown, err := ioutil.ReadAll(r)
	defer r.Close()
	if err != nil {
		return fi, r, err
	}

	// Rename the file to HTML
	fi.Name = fmt.Sprintf("%s.html",
		strings.TrimSuffix(fi.Name, filepath.Ext(fi.Name)))

	html := blackfriday.MarkdownBasic(markdown)

	// Now return it all, with a ReadCloser for the html. Note that
	// the byte array is being returned as a bytes.Reader, with a fake
	// Close() method, via the mutil.ByteCloser() func.
	return fi, mutil.ByteCloser(html), nil
}
