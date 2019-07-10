package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
)

// ApplicationSummary represents an application with its processes and droplet.
type ApplicationSummary struct {
	Application
	CurrentDroplet   Droplet
	ProcessSummaries ProcessSummaries
	Routes           []Route
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
func (actor Actor) GetApplicationSummaryByNameAndSpace(appName string, spaceGUID string, withObfuscatedValues bool) (ApplicationSummary, Warnings, error) {
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
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		if _, ok := err.(actionerror.DropletNotFoundError); !ok {
			return ApplicationSummary{}, allWarnings, err
		}
	}

	var appRoutes []Route
	routes, warnings, err := actor.GetApplicationRoutes(app.GUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		if _, ok := err.(ccerror.ResourceNotFoundError); !ok {
			return ApplicationSummary{}, allWarnings, err
		}
	}
	appRoutes = routes

	summary := ApplicationSummary{
		Application:      app,
		ProcessSummaries: processSummaries,
		CurrentDroplet:   droplet,
		Routes:           appRoutes,
	}
	return summary, allWarnings, nil
}
