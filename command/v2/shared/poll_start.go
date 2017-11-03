package shared

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

func PollStart(ui command.UI, config command.Config, messages <-chan *v2action.LogMessage, logErrs <-chan error, appState <-chan v2action.ApplicationStateChange, apiWarnings <-chan string, apiErrs <-chan error) error {
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
			case actionerror.NOAATimeoutError:
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
			case actionerror.StagingFailedError:
				return translatableerror.StagingFailedError{Message: err.Error()}
			case actionerror.StagingFailedNoAppDetectedError:
				return translatableerror.StagingFailedNoAppDetectedError{BinaryName: config.BinaryName(), Message: err.Error()}
			case actionerror.StagingTimeoutError:
				return translatableerror.StagingTimeoutError{AppName: err.AppName, Timeout: err.Timeout}
			case actionerror.ApplicationInstanceCrashedError:
				return translatableerror.UnsuccessfulStartError{AppName: err.Name, BinaryName: config.BinaryName()}
			case actionerror.ApplicationInstanceFlappingError:
				return translatableerror.UnsuccessfulStartError{AppName: err.Name, BinaryName: config.BinaryName()}
			case actionerror.StartupTimeoutError:
				return translatableerror.StartupTimeoutError{AppName: err.Name, BinaryName: config.BinaryName()}
			default:
				return apiErr
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
