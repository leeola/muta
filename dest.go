package muta

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/leeola/muta/logging"
)

const destPluginName string = "muta.Dest"

type DestOpts struct {
	// Remove the entire destination directory before writing anything.
	//
	// Note: This is entirely handled by the DestOpts() func, not
	// the Stream itself.
	//
	// TODO: Move this option to the DestWithOpts() func, or possibly
	// move the Clean functionality into the Streamer object itself,
	// so that it makes sense to keep this Option here.
	Clean bool

	// Overwrite the contents of any encountered files. If false, an error
	// is returned if any Streamed files exist in the output directory.
	Overwrite bool
}

// Return a DestStream{} Embedder, with the following default options:
//
//		DestOpts{
//			Clean:     false,
//			Overwrite: true,
//		}
func Dest(d string) StreamEmbedder {
	opts := DestOpts{
		Clean:     false,
		Overwrite: true,
	}
	return DestWithOpts(d, opts)
}

// Return a DestStream{} Embedder, with the given options. If
// DestOpts.Clean is true, remove the entire Destination directory before
// creating the Streamer.
func DestWithOpts(d string, opts DestOpts) StreamEmbedder {
	return StreamEmbedderFunc(func(inner Streamer) Streamer {
		if opts.Clean {
			err := os.RemoveAll(d)
			if err != nil {
				return NewErrorStreamer(fmt.Sprintf("%s: %s",
					destPluginName, err.Error()))
			}
		}

		// Make the destination if needed
		if err := os.MkdirAll(d, 0755); err != nil {
			return NewErrorStreamer(fmt.Sprintf("%s: %s",
				destPluginName, err.Error()))
		}

		return &DestStream{
			Streamer:    inner,
			Destination: d,
			Opts:        opts,
		}
	})
}

type DestStream struct {
	// The inner streamer, which provides files for Dest to consume & write
	Streamer

	Destination string
	Opts        DestOpts
}

func (s *DestStream) Use(embedder StreamEmbedder) Streamer {
	return embedder.Embed(s)
}

func (s *DestStream) Next() (*FileInfo, io.ReadCloser, error) {
	// If Dest doesn't have an embedded Streamer do nothing.
	if s.Streamer == nil {
		return nil, nil, nil
	}

	fi, r, err := s.Streamer.Next()

	// If the file is nil, or there's an error, we can't write anything.
	if fi == nil || err != nil {
		return fi, r, err
	}

	destPath := filepath.Join(s.Destination, fi.Path)
	destFilepath := filepath.Join(destPath, fi.Name)

	// MkdirAll checks if the given path is a dir, and exists. If
	// it does not exist, it creates it. So i believe there is no
	// reason for us to bother checking.
	err = os.MkdirAll(destPath, 0755)
	if err != nil {
		return fi, r, err
	}

	logging.Debug([]string{destPluginName}, "Opening", destFilepath)

	var f *os.File
	osFi, err := os.Stat(destFilepath)

	// This area is a bit of a cluster f*ck. In short:
	//
	// 1. If there is an error, and the error is that the file
	// does not exist, create the file.
	// 2. If it's not a file does not exist error, return it.
	// 3. If there is no error, and the filepath is a directory,
	// return an error.
	// 4. If it's not a directory, and we're not allowed to overwrite
	// it, return an error.
	// 5. If we are allowed to overwrite it, open it up.
	//
	// Did i drink too much while writing this? It feels so messy.

	if err != nil {
		// Error opening file

		if os.IsNotExist(err) {
			f, err = os.Create(destFilepath)
			defer f.Close()
			if err != nil {
				// Failed to create file, return
				return fi, r, err
			}
		} else {
			// Stat() error is unknown, return
			return fi, r, err
		}

	} else {
		// No Error opening file

		// There was no error Stating path, it exist
		if osFi.IsDir() {
			// The file path is a dir, return error
			return fi, r, errors.New(fmt.Sprintf(
				"%s: Cannot write to '%s', path is directory.",
				destPluginName,
				destFilepath,
			))
		} else if !s.Opts.Overwrite {
			// We're not allowed to overwrite. Return error.
			return fi, r, errors.New(fmt.Sprintf(
				"%s: Cannot write to '%s', path exists and Overwrite is set "+
					"to false.",
				destPluginName,
				destFilepath,
			))
		} else {
			f, err = os.Create(destFilepath)
			defer f.Close()
			if err != nil {
				// Failed to open file for writing.
				return fi, r, err
			}
		}
	}

	// Finally, copy our reader to our writer (file)
	_, err = io.Copy(f, r)

	return fi, r, nil
}
