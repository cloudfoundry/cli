package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
)

//go:generate counterfeiter . RouteActor

type RouteActor interface {
	GetApplicationRoutes(appGUID string) (v2action.Routes, v2action.Warnings, error)
}

// ApplicationSummary represents an application with its processes and droplet.
type ApplicationSummary struct {
	Application
	CurrentDroplet   Droplet
	ProcessSummaries ProcessSummaries
	Routes           v2action.Routes
}

func (a ApplicationSummary) GetIsolationSegmentName() (string, bool) {
	if a.hasIsolationSegment() {
		return a.ProcessSummaries[0].InstanceDetails[0].IsolationSegment, true
	}
	return "", false
}

func (a ApplicationSummary) hasIsolationSegment() bool {
	return len(a.ProcessSummaries) > 0 &&
		len(a.ProcessSummaries[0].InstanceDetails) > 0 &&
		len(a.ProcessSummaries[0].InstanceDetails[0].IsolationSegment) > 0
}

// GetApplicationSummaryByNameAndSpace returns an application with process and
// instance stats.
func (actor Actor) GetApplicationSummaryByNameAndSpace(appName string, spaceGUID string, withObfuscatedValues bool, routeActor RouteActor) (ApplicationSummary, Warnings, error) {
	app, allWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return ApplicationSummary{}, allWarnings, err
	}

	processSummaries, processWarnings, err := actor.getProcessSummariesForApp(app.GUID, withObfuscatedValues)
	allWarnings = append(allWarnings, processWarnings...)
	if err != nil {
		return ApplicationSummary{}, allWarnings, err
	}

	droplet, warnings, err := actor.GetCurrentDropletByApplication(app.GUID)
	allWarnings = append(allWarnings, Warnings(warnings)...)
	if err != nil {
		if _, ok := err.(actionerror.DropletNotFoundError); !ok {
			return ApplicationSummary{}, allWarnings, err
		}
	}

	var appRoutes v2action.Routes
	if routeActor != nil {
		routes, warnings, err := routeActor.GetApplicationRoutes(app.GUID)
		allWarnings = append(allWarnings, Warnings(warnings)...)
		if err != nil {
			if _, ok := err.(ccerror.ResourceNotFoundError); !ok {
				return ApplicationSummary{}, allWarnings, err
			}
		}
		appRoutes = routes
	}

	summary := ApplicationSummary{
		Application:      app,
		ProcessSummaries: processSummaries,
		CurrentDroplet:   droplet,
		Routes:           appRoutes,
	}
	return summary, allWarnings, nil
}
