//
// # Muta Bin
//
// The muta bin is the main bin called by users. It is responsible for
// finding and running the specified `muta.go` file in the current/target
// directory. Though that logic is actually handled by GoScriptify.
//
package main

import (
	"fmt"
	"os"

	"github.com/leeola/goscriptify"
	"github.com/leeola/muta"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println(fmt.Sprintf("Muta %s (bin)", muta.VERSION))
	}

	// Proxy this bin input/output to the "muta" file
	// in the current directory
	goscriptify.RunOneScript("muta", "Muta", "muta.go", "muta/muta.go", ".muta/muta.go")
}
