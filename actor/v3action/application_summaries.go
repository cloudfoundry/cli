package v3action

import "net/url"

func (actor Actor) GetApplicationSummariesBySpace(spaceGUID string) ([]ApplicationSummary, Warnings, error) {
	var allWarnings Warnings

	apps, warnings, err := actor.CloudControllerClient.GetApplications(url.Values{
		"space_guids": []string{spaceGUID},
	})
	allWarnings = Warnings(warnings)
	if err != nil {
		return nil, allWarnings, err
	}

	var appSummaries []ApplicationSummary

	for _, app := range apps {
		processSummaries, processWarnings, err := actor.getProcessSummariesForApp(app.GUID)
		allWarnings = append(allWarnings, processWarnings...)
		if err != nil {
			return nil, allWarnings, err
		}

		appSummaries = append(appSummaries, ApplicationSummary{
			Application:      Application(app),
			ProcessSummaries: processSummaries,
		})
	}

	return appSummaries, allWarnings, nil
}
