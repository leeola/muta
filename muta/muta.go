//
// # Muta Bin
//
// The muta bin is the main bin called by users. It is responsible for
// finding and running the specified `muta.go` file in the current/target
// directory.
//
package main

import (
	"fmt"
	"os"

	"github.com/leeola/goscriptify"
)

// githash's value is filled in via -ldflags -X.
var githash string = ""

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println(fmt.Sprintf("Muta Bin: %s", githash))
	}

	// Proxy this bin input/output to the "muta" file
	// in the current directory
	goscriptify.RunScript("muta")
}
