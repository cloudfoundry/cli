package uihelpers

import (
	"fmt"
	"strings"

	. "code.cloudfoundry.org/cli/cf/i18n"

	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/terminal"
)

func ColoredAppState(app models.ApplicationFields) string {
	appState := strings.ToLower(app.State)

	if app.RunningInstances == 0 {
		if appState == models.ApplicationStateStopped {
			return appState
		}
		return terminal.CrashedColor(appState)
	}

	if app.RunningInstances < app.InstanceCount {
		return terminal.WarningColor(appState)
	}

	return appState
}

func ColoredAppInstances(app models.ApplicationFields) string {
	healthString := fmt.Sprintf("%d/%d", app.RunningInstances, app.InstanceCount)

	if app.RunningInstances < 0 {
		healthString = fmt.Sprintf("?/%d", app.InstanceCount)
	}

	if app.RunningInstances == 0 {
		if strings.ToLower(app.State) == models.ApplicationStateStopped {
			return healthString
		}
		return terminal.CrashedColor(healthString)
	}

	if app.RunningInstances < app.InstanceCount {
		return terminal.WarningColor(healthString)
	}

	return healthString
}

func ColoredInstanceState(instance models.AppInstanceFields) (colored string) {
	state := string(instance.State)
	switch state {
	case models.ApplicationStateStarted, models.ApplicationStateRunning:
		colored = T("running")
	case models.ApplicationStateStopped:
		colored = terminal.StoppedColor(T("stopped"))
	case models.ApplicationStateCrashed:
		colored = terminal.CrashedColor(T("crashed"))
	case models.ApplicationStateFlapping:
		colored = terminal.CrashedColor(T("crashing"))
	case models.ApplicationStateDown:
		colored = terminal.CrashedColor(T("down"))
	case models.ApplicationStateStarting:
		colored = terminal.AdvisoryColor(T("starting"))
	default:
		colored = terminal.WarningColor(state)
	}

	return
}
