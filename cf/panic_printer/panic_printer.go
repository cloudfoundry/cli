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
