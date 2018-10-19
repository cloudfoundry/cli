package shared

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
)

func PollStage(dropletStream <-chan v7action.Droplet, warningsStream <-chan v7action.Warnings, errStream <-chan error, logStream <-chan *v7action.LogMessage, logErrStream <-chan error, ui command.UI) (v7action.Droplet, error) {
	var closedBuildStream, closedWarningsStream, closedErrStream bool
	var droplet v7action.Droplet

	for {
		select {
		case d, ok := <-dropletStream:
			if !ok {
				closedBuildStream = true
				break
			}
			droplet = d
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

			switch logErr.(type) {
			case actionerror.NOAATimeoutError:
				ui.DisplayWarning("timeout connecting to log server, no log will be shown")
			default:
				ui.DisplayWarning(logErr.Error())
			}
		case err, ok := <-errStream:
			if !ok {
				closedErrStream = true
				break
			}
			return v7action.Droplet{}, err
		}
		if closedBuildStream && closedWarningsStream && closedErrStream {
			return droplet, nil
		}
	}
}
