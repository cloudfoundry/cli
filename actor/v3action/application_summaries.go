package v3action

import (
	"net/url"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

func (actor Actor) GetApplicationSummariesBySpace(spaceGUID string) ([]ApplicationSummary, Warnings, error) {
	var allWarnings Warnings

	apps, warnings, err := actor.CloudControllerClient.GetApplications(url.Values{
		ccv3.SpaceGUIDFilter: []string{spaceGUID},
		ccv3.OrderBy:         []string{ccv3.NameOrder},
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
			Application: Application{
				Name:  app.Name,
				GUID:  app.GUID,
				State: app.State,
				Lifecycle: AppLifecycle{
					Type: AppLifecycleType(app.Lifecycle.Type),
					Data: AppLifecycleData(app.Lifecycle.Data),
				},
			},
			ProcessSummaries: processSummaries,
		})
	}

	return appSummaries, allWarnings, nil
}
