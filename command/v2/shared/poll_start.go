package shared

import (
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

func PollStart(ui command.UI, config command.Config, messages <-chan *v2action.LogMessage, logErrs <-chan error, appState <-chan v2action.ApplicationState, apiWarnings <-chan string, apiErrs <-chan error) error {
	var breakAppState, breakWarnings, breakAPIErrs bool
	for {
		select {
		case message, ok := <-messages:
			if !ok {
				break
			}

			if message.Staging() {
				ui.DisplayLogMessage(message, false)
			}
		case state, ok := <-appState:
			if !ok {
				breakAppState = true
				break
			}

			switch state {
			case v2action.ApplicationStateStopping:
				ui.DisplayNewline()
				ui.DisplayText("Stopping app...")

			case v2action.ApplicationStateStaging:
				ui.DisplayNewline()
				ui.DisplayText("Staging app and tracing logs...")

			case v2action.ApplicationStateStarting:
				ui.DisplayNewline()
				ui.DisplayText("Waiting for app to start...")
			}
		case warning, ok := <-apiWarnings:
			if !ok {
				breakWarnings = true
				break
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
				ui.DisplayWarning(logErr.Error())
			}
		case apiErr, ok := <-apiErrs:
			if !ok {
				breakAPIErrs = true
				break
			}

			switch err := apiErr.(type) {
			case v2action.StagingFailedError:
				return translatableerror.StagingFailedError{Message: err.Error()}
			case v2action.StagingFailedNoAppDetectedError:
				return translatableerror.StagingFailedNoAppDetectedError{BinaryName: config.BinaryName(), Message: err.Error()}
			case v2action.StagingTimeoutError:
				return translatableerror.StagingTimeoutError{AppName: err.Name, Timeout: err.Timeout}
			case v2action.ApplicationInstanceCrashedError:
				return translatableerror.UnsuccessfulStartError{AppName: err.Name, BinaryName: config.BinaryName()}
			case v2action.ApplicationInstanceFlappingError:
				return translatableerror.UnsuccessfulStartError{AppName: err.Name, BinaryName: config.BinaryName()}
			case v2action.StartupTimeoutError:
				return translatableerror.StartupTimeoutError{AppName: err.Name, BinaryName: config.BinaryName()}
			default:
				return HandleError(apiErr)
			}
		}

		// only wait for non-nil channels to be closed
		if (appState == nil || breakAppState) &&
			(apiWarnings == nil || breakWarnings) &&
			(apiErrs == nil || breakAPIErrs) {
			return nil
		}
	}
}
