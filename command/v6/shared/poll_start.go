package shared

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

func PollStart(ui command.UI, config command.Config, messages <-chan sharedaction.LogMessage, logErrs <-chan error, appState <-chan v2action.ApplicationStateChange, apiWarnings <-chan string, apiErrs <-chan error) (apiError error) {
	for appState != nil || apiWarnings != nil || apiErrs != nil  {
		select {
		case message, ok := <-messages:
			if !ok {
				messages = nil
				continue
			}

			if message.Staging() {
				ui.DisplayLogMessage(message, false)
			}

		case state, ok := <-appState:
			if !ok {
				appState = nil
				continue
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
				apiWarnings = nil
				continue
			}

			ui.DisplayWarning(warning)

		case logErr, ok := <-logErrs:
			if !ok {
				logErrs = nil
				continue
			}

			switch logErr.(type) {
			case actionerror.NOAATimeoutError:
				ui.DisplayWarning("timeout connecting to log server, no log will be shown")
			default:
				ui.DisplayWarning(logErr.Error())
			}

		case e, ok := <-apiErrs:
			if !ok {
				apiErrs = nil
				continue
			}
			switch err := e.(type) {
			case actionerror.StagingFailedError:
				apiError = translatableerror.StagingFailedError{Message: err.Error()}
			case actionerror.StagingFailedNoAppDetectedError:
				apiError = translatableerror.StagingFailedNoAppDetectedError{BinaryName: config.BinaryName(), Message: err.Error()}
			case actionerror.StagingTimeoutError:
				apiError = translatableerror.StagingTimeoutError{AppName: err.AppName, Timeout: err.Timeout}
			case actionerror.ApplicationInstanceCrashedError:
				apiError = translatableerror.ApplicationUnableToStartError{AppName: err.Name, BinaryName: config.BinaryName()}
			case actionerror.ApplicationInstanceFlappingError:
				apiError = translatableerror.ApplicationUnableToStartError{AppName: err.Name, BinaryName: config.BinaryName()}
			case actionerror.StartupTimeoutError:
				apiError = translatableerror.StartupTimeoutError{AppName: err.Name, BinaryName: config.BinaryName()}
			default:
				apiError = err
			}
			// if an api error occurred, exit immediately
			return apiError
		}
	}
	return nil
}
