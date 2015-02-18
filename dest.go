package muta

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type DestOpts struct {
	// Not implemented
	Clean bool
	// Not implemented
	Overwrite bool
}

func Dest(d string, args ...interface{}) Streamer {
	var opts DestOpts
	if len(args) == 0 {
		opts = DestOpts{
			Clean:     false,
			Overwrite: true,
		}
	} else if len(args) == 1 {
		_opts, ok := args[0].(DestOpts)
		opts = _opts
		if !ok {
			return ErrorStreamer(errors.New(
				"Unrecognized type in Dest(string, ...interface{}). " +
					"Use DestOpts()",
			))
		}
	}

	if opts.Clean {
		err := os.RemoveAll(d)
		if err != nil {
			return ErrorStreamer(err)
		}
	}

	// Make the destination if needed
	if err := os.MkdirAll(d, 0755); err != nil {
		return ErrorStreamer(err)
	}

	// A staging variable for the currently working file.
	var f *os.File
	return func(fi *FileInfo, chunk []byte) (*FileInfo,
		[]byte, error) {

		// If fi is nil, then this func is now the generator. Dest() has no
		// need to generate, so signal EOS
		if fi == nil {
			return nil, chunk, nil
		}

		// If chunk is nil, we're at EOF
		if chunk == nil {
			var err error
			// Close f, and set it to nil if needed
			if f != nil {
				err = f.Close()
				f = nil
			}
			return nil, nil, err
		}

		destPath := filepath.Join(d, fi.Path)
		destFilepath := filepath.Join(destPath, fi.Name)

		// If f is nil, we're at a new file
		if f == nil {
			// MkdirAll checks if the given path is a dir, and exists. So
			// i believe there is no reason for us to bother checking.
			err := os.MkdirAll(destPath, 0755)
			if err != nil {
				return fi, chunk, err
			}

			osFi, err := os.Stat(destFilepath)
			if err == nil && osFi.IsDir() {
				return fi, chunk, errors.New(fmt.Sprintf(
					"Cannot write to '%s', path is directory.",
					destFilepath,
				))
			}

			// This area is a bit of a cluster f*ck. In short:
			//
			// 1. If there is an error, and the error is that the file
			// does not exist, create it.
			// 2. If it's not a file does not exist error, return it.
			// 3. If there is no error, and the filepath is a directory,
			// return an error.
			// 4. If it's not a directory, and we're not allowed to overwrite
			// it, return an error.
			// 5. If we are allowed to overwrite it, open it up.
			//
			// Did i drink too much while writing this? It feels so messy.
			if err != nil {
				if os.IsNotExist(err) {
					f, err = os.Create(destFilepath)
					if err != nil {
						// Failed to create file, return
						return fi, chunk, err
					}
				} else {
					// Stat() error is unknown, return
					return fi, chunk, err
				}
			} else {
				// There was no error Stating path, it exist
				if osFi.IsDir() {
					// The file path is a dir, return error
					return fi, chunk, errors.New(fmt.Sprintf(
						"Cannot write to '%s', path is directory.",
						destFilepath,
					))
				} else if !opts.Overwrite {
					// We're not allowed to overwrite. Return error.
					return fi, chunk, errors.New(fmt.Sprintf(
						"Cannot write to '%s', path exists and Overwrite is set "+
							"to false.",
						destFilepath,
					))
				} else {
					f, err = os.Open(destFilepath)
					if err != nil {
						// Failed to open file for writing.
						return fi, chunk, err
					}
				}
			}
		}

		// lenth written can be ignored, because Write() returns an error
		// if len(chunk) != n
		_, err := f.Write(chunk)

		// Return EOS always. Dest() writes everything, like a boss..?
		return nil, nil, err
	}
}
