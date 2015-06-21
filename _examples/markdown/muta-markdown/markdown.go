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

func Markdown() muta.Streamer {
	return &MarkdownStreamer{}
}

type MarkdownStreamer struct {
}

// The Next() method is the (only) workhorse of a Streamer. The Stream
// will call it with a FileInfo and a ReadCloser, expecting the Next()
// method to modify them as it sees fit.
//
// If no FileInfo is provided, the Next() method is expected to create
// new files and return them - or return nothing and not be called again.
//
// Next() will always be called once more if it returns a file, unless
// it returns an error.
func (s *MarkdownStreamer) Next(fi muta.FileInfo, rc io.ReadCloser) (
	muta.FileInfo, io.ReadCloser, error) {

	// MarkdownStreamer does not create any files, so if no files are
	// given to it, just return.
	if fi == nil {
		return fi, rc, nil
	}

	// If the file isn't markdown, we don't care about it. Return it
	// unmodified.
	if filepath.Ext(fi.Name()) != ".md" {
		return fi, rc, nil
	}

	// Since the file is markdown, read it all so we can convert it to
	// markdown.
	markdown, err := ioutil.ReadAll(rc)
	defer rc.Close()
	if err != nil {
		return fi, rc, err
	}

	// Rename the file to HTML
	fi.SetName(fmt.Sprintf("%s.html",
		strings.TrimSuffix(fi.Name(), filepath.Ext(fi.Name())),
	))

	// Use Blackfriday to create our Markdown
	html := blackfriday.MarkdownBasic(markdown)

	// ByteCloser() is a muta utility function that takes a byte array, and
	// returns a fake ReadCloser. This is needed to satisfy the Streamer
	// interface.
	rc = mutil.ByteCloser(html)

	// Now return it all, for subsequent plugins to modify, write to file, etc.
	return fi, rc, nil
}
