package shared

import (
	"time"

	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/clock"
)

func PollStage(appName string, buildStream <-chan v3action.Build, warningsStream <-chan v3action.Warnings, errStream <-chan error, logStream <-chan *v3action.LogMessage, logErrStream <-chan error, ui command.UI, clock clock.Clock, stagingTimeout time.Duration) (string, error) {
	var closedBuildStream, closedWarningsStream, closedErrStream bool
	var dropletGUID string

	timer := clock.NewTimer(stagingTimeout)

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
			return "", HandleError(err)
		case <-timer.C():
			return "", translatableerror.StagingTimeoutError{
				AppName: appName,
				Timeout: stagingTimeout,
			}
		}
		if closedBuildStream && closedWarningsStream && closedErrStream {
			return dropletGUID, nil
		}
	}
}
