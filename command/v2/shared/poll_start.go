package shared

import (
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
)

func PollStart(ui command.UI, config command.Config, messages <-chan *v2action.LogMessage, logErrs <-chan error, apiWarnings <-chan string, apiErrs <-chan error) error {
	for {
		select {
		case message, ok := <-messages:
			if !ok {
				break
			}

			ui.DisplayLogMessage(message, false)
		case warning, ok := <-apiWarnings:
			if !ok {
				return nil
			}

			ui.DisplayWarning(warning)
		case logErr, ok := <-logErrs:
			if !ok {
				break
			}

			switch logErr.(type) {
			case v2action.NOAATimeoutError:
				ui.DisplayWarning("timeout connecting to log server, no log will be shown")
			default:
				return HandleError(logErr)
			}
		case apiErr, ok := <-apiErrs:
			if !ok {
				return nil
			}

			if stgErr, ok := apiErr.(v2action.StagingFailedError); ok {
				return StagingFailedError{
					BinaryName: config.BinaryName(),
					Message:    stgErr.Error(),
				}
			}

			return HandleError(apiErr)
		}
	}
}
