package v3action

import "code.cloudfoundry.org/cli/actor/actionerror"

// ApplicationSummary represents an application with its processes and droplet.
type ApplicationSummary struct {
	Application
	ProcessSummaries ProcessSummaries
	CurrentDroplet   Droplet
}

// GetApplicationSummaryByNameAndSpace returns an application with process and
// instance stats.
func (actor Actor) GetApplicationSummaryByNameAndSpace(appName string,
	spaceGUID string) (ApplicationSummary, Warnings, error) {
	app, allWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return ApplicationSummary{}, allWarnings, err
	}

	processSummaries, processWarnings, err := actor.getProcessSummariesForApp(app.GUID)
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

	summary := ApplicationSummary{
		Application:      app,
		ProcessSummaries: processSummaries,
		CurrentDroplet:   droplet,
	}
	return summary, allWarnings, nil
}
