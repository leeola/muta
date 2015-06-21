//
// # Muta Gofile
//
package main

import (
	markdown "./muta-markdown"
	"github.com/leeola/muta"
)

func main() {
	// Add our markdown task
	muta.Task("markdown", func() muta.Stream {
		// Create a Stream with Src(), and load files matching *.md
		// into the Stream
		s := muta.Src("./*.md").
			// Pipe all of the Src() files into our local Markdown plugin.
			Pipe(markdown.Markdown()).
			// Pipe the output of Markdown() into the Dest() Streamer, which
			// will write all incoming files info the build directory
			Pipe(muta.Dest("./build"))

		// Return our Stream. Muta will start the Stream once the
		// Task is executed.
		return s
	})

	muta.Task("default", "markdown")
	muta.Te()
}
