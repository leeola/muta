//
// # Muta Bin
//
// Handle the CLI in/out of a `Muta` file that was ran.
//
package main

import "github.com/docopt/docopt-go"

func ParseArgs() {
	usage := `Muta(te)

Usage:
	muta [task]
  muta -h | --help
  muta --version

Options:
  -h --help     Show this screen.
  --version     Show version.`

	docopt.Parse(usage, nil, true, "Muta 0.0.0", false)
}

func main() {
}
