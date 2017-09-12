package v3action

import "net/url"

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

	var droplet Droplet
	ccv3Droplets, warnings, err := actor.CloudControllerClient.GetApplicationDroplets(
		app.GUID,
		url.Values{"current": []string{"true"}},
	)
	allWarnings = append(allWarnings, Warnings(warnings)...)
	if err != nil {
		return ApplicationSummary{}, allWarnings, err
	}

	if len(ccv3Droplets) == 1 {
		droplet.Stack = ccv3Droplets[0].Stack
		droplet.Image = ccv3Droplets[0].Image
		for _, ccv3Buildpack := range ccv3Droplets[0].Buildpacks {
			droplet.Buildpacks = append(droplet.Buildpacks, Buildpack(ccv3Buildpack))
		}
	}

	summary := ApplicationSummary{
		Application:      app,
		ProcessSummaries: processSummaries,
		CurrentDroplet:   droplet,
	}
	return summary, allWarnings, nil
}
