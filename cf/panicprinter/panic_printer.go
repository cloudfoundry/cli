package panicprinter

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/version"
)

type PanicPrinter struct {
	ui terminal.UI
}

func NewPanicPrinter(ui terminal.UI) PanicPrinter {
	return PanicPrinter{
		ui: ui,
	}
}

func (p PanicPrinter) DisplayCrashDialog(err interface{}, commandArgs string, stackTrace string) {
	if err != nil {
		switch err := err.(type) {
		case errors.Exception:
			p.ui.Say(CrashDialog(err.Message, commandArgs, stackTrace))
		case error:
			p.ui.Say(CrashDialog(err.Error(), commandArgs, stackTrace))
		case string:
			p.ui.Say(CrashDialog(err, commandArgs, stackTrace))
		default:
			p.ui.Say(CrashDialog("An unexpected type of error", commandArgs, stackTrace))
		}
	}
}

func CrashDialog(errorMessage string, commandArgs string, stackTrace string) string {
	formattedString := `
	Something unexpected happened. This is a bug in %s.

	Please re-run the command that caused this exception with the environment
	variable CF_TRACE set to true.

	Also, please update to the latest cli and try the command again:
	https://code.cloudfoundry.org/cli/releases

	Please create an issue at: https://code.cloudfoundry.org/cli/issues

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

	return fmt.Sprintf(formattedString, cf.Name, commandArgs, version.BinaryVersion, errorMessage, stackTrace)
}
