package v2action

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"

type ApplicationSummary struct {
	Application
	Stack            Stack
	RunningInstances []ApplicationInstance
	Routes           []Route
}

func (actor Actor) GetApplicationSummaryByNameAndSpace(name string, spaceGUID string) (ApplicationSummary, Warnings, error) {
	var allWarnings Warnings

	app, warnings, err := actor.GetApplicationByNameAndSpace(name, spaceGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return ApplicationSummary{}, allWarnings, err
	}

	applicationSummary := ApplicationSummary{
		Application: app,
	}

	if app.State != ccv2.ApplicationStopped {
		var instances []ApplicationInstance
		instances, warnings, err = actor.GetApplicationInstancesByApplication(app.GUID)
		allWarnings = append(allWarnings, warnings...)

		switch err.(type) {
		case nil:
			applicationSummary.RunningInstances = instances
		case ccv2.AppStoppedStatsError:
			// Don't do anything
		default:
			return ApplicationSummary{}, allWarnings, err
		}
	}

	routes, warnings, err := actor.GetApplicationRoutes(app.GUID, nil)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return ApplicationSummary{}, allWarnings, err
	}
	applicationSummary.Routes = routes

	stack, warnings, err := actor.GetStack(app.StackGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return ApplicationSummary{}, allWarnings, err
	}
	applicationSummary.Stack = stack

	return applicationSummary, allWarnings, nil
}
