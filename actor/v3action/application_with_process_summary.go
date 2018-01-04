package v3action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
)

type ApplicationWithProcessSummary struct {
	Application
	ProcessSummaries ProcessSummaries
}

func (actor Actor) GetApplicationsWithProcessesBySpace(spaceGUID string) ([]ApplicationWithProcessSummary, Warnings, error) {
	var allWarnings Warnings

	apps, warnings, err := actor.CloudControllerClient.GetApplications(
		ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{spaceGUID}},
		ccv3.Query{Key: ccv3.OrderBy, Values: []string{ccv3.NameOrder}},
	)
	allWarnings = Warnings(warnings)
	if err != nil {
		return nil, allWarnings, err
	}

	var appSummaries []ApplicationWithProcessSummary

	for _, app := range apps {
		processSummaries, processWarnings, err := actor.getProcessSummariesForApp(app.GUID)
		allWarnings = append(allWarnings, processWarnings...)
		if err != nil {
			return nil, allWarnings, err
		}

		appSummaries = append(appSummaries, ApplicationWithProcessSummary{
			Application: Application{
				Name:  app.Name,
				GUID:  app.GUID,
				State: app.State,
				Lifecycle: AppLifecycle{
					Type: constant.AppLifecycleType(app.Lifecycle.Type),
					Data: AppLifecycleData(app.Lifecycle.Data),
				},
			},
			ProcessSummaries: processSummaries,
		})
	}

	return appSummaries, allWarnings, nil
}
