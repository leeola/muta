//
// # Muta Bin
//
// Handle the CLI in/out of a `Muta` file that was ran.
//
package muta

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/docopt/docopt-go"
)

func ParseArgs(tasks []string) map[string]interface{} {
	sTasks := ""
	if tasks != nil && len(tasks) > 0 {
		sTasks = fmt.Sprintf(`
Tasks:
  %s
`, strings.Join(tasks, "\n  "))
	}

	usage := fmt.Sprintf(`Muta(te)

Usage:
  muta [<task>]
  muta -h | --help
  muta --version
%s
Options:
  -h --help     Show this screen.
  --version     Show version.`, sTasks)

	args, _ := docopt.Parse(
		usage, nil, true, fmt.Sprintf("Muta %s (lib)", VERSION), false,
	)

	return args
}

func Te() {
	taskNames := []string{}
	for tn := range DefaultTasker.Tasks {
		taskNames = append(taskNames, tn)
	}

	args := ParseArgs(taskNames)

	var err error
	if args["<task>"] == nil {
		err = DefaultTasker.Run()
	} else {
		name, ok := args["<task>"].(string)
		if ok {
			err = DefaultTasker.RunTask(name)
		} else {
			err = errors.New("<task> was not a string")
		}
	}

	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
