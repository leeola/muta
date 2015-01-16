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

// The dynamic use of task names makes this function a bit of a
// clusterfuck. This needs to be cleaned up, along with the
// taskNames logic below.
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
	// TODO: This taskNames logic needs to be separated and improved.
	// I'm thinking, it's own func, and return a formatted docopt friendly
	// string
	taskNames := []string{}
	defaultTaskNames := []string{}
	if DefaultTasker.Tasks["default"] != nil {
		defaultTaskNames = DefaultTasker.Tasks["default"].Dependencies
	}
	for tn := range DefaultTasker.Tasks {
		if tn == "default" {
			continue
		}
		// A dirty implementation of default appending.
		for _, dtn := range defaultTaskNames {
			// If the taskName is a dependency of the Default Task, append
			// " (default)" to the taskName
			if tn == dtn {
				tn += " (default)"
				break
			}
		}
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
