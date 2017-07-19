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

	var summaries []ApplicationSummary

	for _, app := range apps {
		processes, processWarnings, err := actor.getProcessesForApp(app.GUID)
		allWarnings = append(allWarnings, processWarnings...)
		if err != nil {
			return nil, allWarnings, err
		}

		summaries = append(summaries, ApplicationSummary{
			Application: Application(app),
			Processes:   processes,
		})
	}

	return summaries, allWarnings, nil
}
