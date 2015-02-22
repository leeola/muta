package mutamarkdown

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/leeola/muta"
	"github.com/russross/blackfriday"
)

// Called when our Streamer is the Generator
func generate() (*muta.FileInfo, []byte, error) {
	return nil, nil, nil
}

// Called when data is coming in
func buffer(b *bytes.Buffer, chunk []byte) (*muta.FileInfo, []byte, error) {
	_, err := b.Write(chunk)
	// Note that by returning `nil` File, we signal "End of Stream" (EOS).
	// This causes the Stream to not call any Streamers *after* this
	// stream.
	//
	// We do this because we want to buffer all of the incoming data for
	// each file. Once we collect it all, we modify it, and then return
	// it.
	return nil, nil, err
}

// The incoming data stream for the given file is done, we can
// compile the markdown and return our modified data
func write(b *bytes.Buffer, fi *muta.FileInfo, _ []byte) (*muta.FileInfo, []byte, error) {
	if filepath.Ext(fi.Name) != ".html" {
		fi.Name = strings.Replace(fi.Name, filepath.Ext(fi.Name), ".html", 1)
	}

	// If there is no data to write, call EOF
	if b.Len() == 0 {
		return fi, nil, nil
	}

	rawMarkdown, err := ioutil.ReadAll(b)
	if err != nil {
		return fi, nil, err
	}

	html := blackfriday.MarkdownBasic(rawMarkdown)
	return fi, html, nil
}

func Markdown() muta.Streamer {
	var b bytes.Buffer
	return muta.NewEasyStreamer("./markdown.Markdown", func(fi *muta.FileInfo, chunk []byte) (*muta.FileInfo, []byte, error) {
		switch {
		case fi == nil:
			// If fi is nil, Markdown() is being asked to generate files.
			return generate()

		case filepath.Ext(fi.Name) != ".md":
			// If the file is not Markdown, pass it through untouched.
			return fi, chunk, nil

		case chunk == nil:
			// If chunk is nil, we're at the EOF for the incoming data for *fi
			return write(&b, fi, chunk)

		default:
			// If chunk isn't nil, buffer the data
			return buffer(&b, chunk)
		}
	})
}
