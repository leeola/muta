//
// # Muta Bin
//
// Handle the CLI in/out of a `Muta` file that was ran.
//
package muta

import (
	"fmt"
	"strings"

	"github.com/docopt/docopt-go"
)

func ParseArgs(tasks []string) {
	sTasks := ""
	if tasks != nil && len(tasks) > 0 {
		sTasks = fmt.Sprintf(`
Tasks:
  %s
`, strings.Join(tasks, "\n  "))
	}

	usage := fmt.Sprintf(`Muta(te)

Usage:
  muta [task]
  muta -h | --help
  muta --version
%s
Options:
  -h --help     Show this screen.
  --version     Show version.`, sTasks)

	docopt.Parse(usage, nil, true, "Muta 0.0.0", false)
}

func Te() {
	ParseArgs([]string{})
}
