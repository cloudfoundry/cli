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
				printCrashDialog(err.Message, commandArgs, stackTrace)
			} else {
				fmt.Println(err.Message)
			}
		case error:
			printCrashDialog(err.Error(), commandArgs, stackTrace)
		case string:
			printCrashDialog(err, commandArgs, stackTrace)
		default:
			printCrashDialog("An unexpected type of error", commandArgs, stackTrace)
		}
	}
}

func CrashDialog(errorMessage string, commandArgs string, stackTrace string) string {
	formattedString := `

	Aww shucks.

	Something completely unexpected happened. This is a bug in %s.
	Please file this bug : https://github.com/cloudfoundry/cli/issues
	Tell us that you ran this command:

		%s

	using this version of the CLI:

		%s

	and that this error occurred:

		%s

	and this stack trace:

	%s
`

	return fmt.Sprintf(formattedString, cf.Name(), commandArgs, cf.Version, errorMessage, stackTrace)
}

func printCrashDialog(errorMessage string, commandArgs string, stackTrace string) {
	UI.Say(CrashDialog(errorMessage, commandArgs, stackTrace))
}
