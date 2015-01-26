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
	muta.Task("markdown", func() (*muta.Stream, error) {
		// Create a Stream with Src(), and load files matching *.md
		// into the Stream
		s := muta.Src("./*.md").
			// Pipe all of the Src() files into our local Markdown plugin.
			Pipe(markdown.Markdown()).
			// Pipe the output of Markdown into the Dest() func, which will
			// write all incoming files info the build directory
			Pipe(muta.Dest("./build"))

		// Return our Stream. Muta will start the Stream once the
		// Task is executed. We could of course, Start the Stream ourselves
		// instead if we so desired.
		return s, nil
	})

	muta.Task("default", "markdown")
	muta.Te()
}
