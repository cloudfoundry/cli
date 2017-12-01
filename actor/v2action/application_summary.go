package v2action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

type ApplicationSummary struct {
	Application
	Stack            Stack
	IsolationSegment string
	RunningInstances []ApplicationInstanceWithStats
	Routes           []Route
}

func (app ApplicationSummary) StartingOrRunningInstanceCount() int {
	count := 0
	for _, instance := range app.RunningInstances {
		if instance.State == ApplicationInstanceState(ccv2.ApplicationInstanceStarting) ||
			instance.State == ApplicationInstanceState(ccv2.ApplicationInstanceRunning) {
			count++
		}
	}
	return count
}

func (actor Actor) GetApplicationSummaryByNameAndSpace(name string, spaceGUID string) (ApplicationSummary, Warnings, error) {
	var allWarnings Warnings

	app, warnings, err := actor.GetApplicationByNameAndSpace(name, spaceGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return ApplicationSummary{}, allWarnings, err
	}

	applicationSummary := ApplicationSummary{Application: app}

	// cloud controller calls the instance reporter only when the desired
	// application state is STARTED
	if app.State == ccv2.ApplicationStarted {
		var instances []ApplicationInstanceWithStats
		instances, warnings, err = actor.GetApplicationInstancesWithStatsByApplication(app.GUID)
		allWarnings = append(allWarnings, warnings...)

		switch err.(type) {
		case nil:
			applicationSummary.RunningInstances = instances

			if len(instances) > 0 {
				applicationSummary.IsolationSegment = instances[0].IsolationSegment
			}
		case actionerror.ApplicationInstancesNotFoundError:
			// don't set instances in summary
		default:
			return ApplicationSummary{}, allWarnings, err
		}
	}

	routes, warnings, err := actor.GetApplicationRoutes(app.GUID)
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
