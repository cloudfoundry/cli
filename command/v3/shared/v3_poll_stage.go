package shared

import (
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
)

func PollStage(buildStream <-chan v3action.Build, warningsStream <-chan v3action.Warnings, errStream <-chan error, logStream <-chan *v3action.LogMessage, logErrStream <-chan error, ui command.UI) (error, string) {

	var closedBuildStream, closedWarningsStream, closedErrStream bool
	var dropletGUID string
	for {
		select {
		case build, ok := <-buildStream:
			if !ok {
				closedBuildStream = true
				break
			}
			dropletGUID = build.Droplet.GUID
			ui.DisplayNewline()
			ui.DisplayText("droplet: {{.DropletGUID}}", map[string]interface{}{"DropletGUID": dropletGUID})
		case log, ok := <-logStream:
			if !ok {
				break
			}
			if log.Staging() {
				ui.DisplayLogMessage(log, false)
			}
		case warnings, ok := <-warningsStream:
			if !ok {
				closedWarningsStream = true
				break
			}
			ui.DisplayWarnings(warnings)
		case logErr, ok := <-logErrStream:
			if !ok {
				break
			}
			ui.DisplayWarning(logErr.Error())
		case err, ok := <-errStream:
			if !ok {
				closedErrStream = true
				break
			}
			return HandleError(err), ""
		}
		if closedBuildStream && closedWarningsStream && closedErrStream {
			return nil, dropletGUID
		}
	}
}
