package panic_printer

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/terminal"
)

var UI terminal.UI

func DisplayCrashDialog(err interface{}, commandArgs string, stackTrace string) {
	if err != nil && err != terminal.QuietPanic {
		switch err := err.(type) {
		case errors.Exception:
			if err.DisplayCrashDialog {
				UI.Say(CrashDialog(err.Message, commandArgs, stackTrace))
			} else {
				fmt.Println(err.Message)
			}
		case error:
			UI.Say(CrashDialog(err.Error(), commandArgs, stackTrace))
		case string:
			UI.Say(CrashDialog(err, commandArgs, stackTrace))
		default:
			UI.Say(CrashDialog("An unexpected type of error", commandArgs, stackTrace))
		}
	}
}

func CrashDialog(errorMessage string, commandArgs string, stackTrace string) string {
	formattedString := `
	Something unexpected happened. This is a bug in %s.

	Please re-run the command that caused this exception with the environment
	variable CF_TRACE set to true.

	Also, please update to the latest cli and try the command again:
	https://github.com/cloudfoundry/cli/releases

	Please create an issue at: https://github.com/cloudfoundry/cli/issues

	Include the below information when creating the issue:

		Command
		%s

		CLI Version
		%s

		Error
		%s

		Stack Trace
		%s

		Your Platform Details
		e.g. Mac OS X 10.11, Windows 8.1 64-bit, Ubuntu 14.04.3 64-bit

		Shell
		e.g. Terminal, iTerm, Powershell, Cygwin, gnome-terminal, terminator
`

	return fmt.Sprintf(formattedString, cf.Name(), commandArgs, cf.Version, errorMessage, stackTrace)
}
