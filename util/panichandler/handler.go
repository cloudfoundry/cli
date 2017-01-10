package panichandler

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"text/template"

	"code.cloudfoundry.org/cli/version"
)

const maxStackSizeLimit = 1024 * 1024

// HandlePanic will recover from any panics and display a friendly error
// message with additional information used for debugging the panic.
func HandlePanic() {
	stackTraceBytes := make([]byte, maxStackSizeLimit)
	runtime.Stack(stackTraceBytes, true)
	stackTrace := "\t" + strings.Replace(string(stackTraceBytes), "\n", "\n\t", -1)
	if err := recover(); err != nil {
		formattedString := `
		Something unexpected happened. This is a bug in {{.Binary}}.

		Please re-run the command that caused this exception with the environment
		variable CF_TRACE set to true.

		Also, please update to the latest cli and try the command again:
		https://code.cloudfoundry.org/cli/releases

		Please create an issue at: https://code.cloudfoundry.org/cli/issues

		Include the below information when creating the issue:

		Command
		{{.Command}}

		CLI Version
		{{.Version}}

		Error
		{{.Error}}

		Stack Trace
		{{.StackTrace}}

		Your Platform Details
		e.g. Mac OS X 10.11, Windows 8.1 64-bit, Ubuntu 14.04.3 64-bit

		Shell
		e.g. Terminal, iTerm, Powershell, Cygwin, gnome-terminal, terminator
		`
		formattedTemplate := template.Must(template.New("Panic Template").Parse(formattedString))
		templateErr := formattedTemplate.Execute(os.Stderr, map[string]interface{}{
			"Binary":     os.Args[0],
			"Command":    strings.Join(os.Args, " "),
			"Version":    version.VersionString(),
			"StackTrace": stackTrace,
			"Error":      err,
		})
		if templateErr != nil {
			fmt.Fprintf(os.Stderr,
				"Unable to format panic response: %s\n",
				templateErr.Error(),
			)

			fmt.Fprintf(os.Stderr,
				"Version:%s\nCommand:%s\nOriginal Stack Trace:%s\nOriginal Error:%s\n",
				version.VersionString(),
				strings.Join(os.Args, " "),
				stackTrace,
				err,
			)
		}
		os.Exit(1)
	}
}
