//
// # Muta Bin
//
// Handle the CLI in/out of a `Muta` file that was ran.
//
package muta

import (
	"fmt"
	"os"
	"strings"

	"github.com/docopt/docopt-go"
	"github.com/leeola/muta/logging"
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
  muta [-l=<level>] [-t=<tags>] [<task>]
  muta -h | --help
  muta --version
%s
Options:
  -l=<level>  The log level [default: info]
  -t=<tags>   A comma separated list of logging tags
  -h --help   Show this screen.
  --version   Show version.
`, sTasks)

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

	// Set logging tags
	if args["-t"] != nil {
		// Don't think Docopt will return anything but a string
		tags, _ := args["-t"].(string)
		logging.SetTags(strings.Split(tags, ",")...)
	} else {
		logging.SetTags()
	}

	// Set logging level
	// (it has a default, so it should never be nil)
	// Don't think Docopt will return anything but a string
	logLevel, _ := args["-l"].(string)
	logging.SetLevel(logging.LevelFromString(logLevel))

	var err error
	if args["<task>"] == nil {
		err = DefaultTasker.Run()
	} else {
		// Don't think Docopt will return anything but a string
		name, _ := args["<task>"].(string)
		err = DefaultTasker.RunTask(name)
	}

	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

// An alias for Te()
func Start() {
	Te()
}
