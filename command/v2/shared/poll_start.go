package shared

import (
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
)

func PollStart(ui command.UI, config command.Config, messages <-chan *v2action.LogMessage, logErrs <-chan error, appStarting <-chan bool, apiWarnings <-chan string, apiErrs <-chan error) error {
	for {
		select {
		case message, ok := <-messages:
			if !ok {
				break
			}

			if message.Staging() {
				ui.DisplayLogMessage(message, false)
			}
		case _, ok := <-appStarting:
			if !ok {
				return nil
			}

			ui.DisplayNewline()
			ui.DisplayText("Waiting for app to start...")
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

			switch err := apiErr.(type) {
			case v2action.StagingFailedError:
				return StagingFailedError{BinaryName: config.BinaryName(), Message: err.Error()}
			case v2action.StagingTimeoutError:
				return StagingTimeoutError{AppName: err.Name, Timeout: err.Timeout}
			case v2action.ApplicationInstanceCrashedError:
				return UnsuccessfulStartError{AppName: err.Name, BinaryName: config.BinaryName()}
			case v2action.ApplicationInstanceFlappingError:
				return UnsuccessfulStartError{AppName: err.Name, BinaryName: config.BinaryName()}
			case v2action.StartupTimeoutError:
				return StartupTimeoutError{AppName: err.Name, BinaryName: config.BinaryName()}
			}

			return HandleError(apiErr)
		}
	}
}
