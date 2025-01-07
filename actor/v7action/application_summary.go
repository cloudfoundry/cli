package v7action

import (
	"errors"

	"code.cloudfoundry.org/cli/v8/actor/actionerror"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v8/resources"
	"code.cloudfoundry.org/cli/v8/util/batcher"
)

type ApplicationSummary struct {
	resources.Application
	ProcessSummaries ProcessSummaries
	Routes           []resources.Route
}

// v7action.DetailedApplicationSummary represents an application with its processes and droplet.
type DetailedApplicationSummary struct {
	ApplicationSummary
	CurrentDroplet resources.Droplet
	Deployment     resources.Deployment
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

func (actor Actor) GetAppSummariesForSpace(spaceGUID string, labelSelector string, omitStats bool) ([]ApplicationSummary, Warnings, error) {
	var allWarnings Warnings
	var allSummaries []ApplicationSummary

	keys := []ccv3.Query{
		{Key: ccv3.SpaceGUIDFilter, Values: []string{spaceGUID}},
		{Key: ccv3.OrderBy, Values: []string{ccv3.NameOrder}},
		{Key: ccv3.PerPage, Values: []string{ccv3.MaxPerPage}},
	}
	if len(labelSelector) > 0 {
		keys = append(keys, ccv3.Query{Key: ccv3.LabelSelectorFilter, Values: []string{labelSelector}})
	}
	apps, ccv3Warnings, err := actor.CloudControllerClient.GetApplications(keys...)
	allWarnings = append(allWarnings, ccv3Warnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	var processSummariesByAppGUID map[string]ProcessSummaries
	var warnings Warnings

	if !omitStats {
		processSummariesByAppGUID, warnings, err = actor.getProcessSummariesForApps(apps)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return nil, allWarnings, err
		}
	}

	var routes []resources.Route

	ccv3Warnings, err = batcher.RequestByGUID(toAppGUIDs(apps), func(guids []string) (ccv3.Warnings, error) {
		batch, warnings, err := actor.CloudControllerClient.GetRoutes(ccv3.Query{
			Key: ccv3.AppGUIDFilter, Values: guids,
		})
		routes = append(routes, batch...)
		return warnings, err
	})
	allWarnings = append(allWarnings, ccv3Warnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	routesByAppGUID := make(map[string][]resources.Route)

	for _, route := range routes {
		for _, dest := range route.Destinations {
			routesByAppGUID[dest.App.GUID] = append(routesByAppGUID[dest.App.GUID], route)
		}
	}

	for _, app := range apps {
		processSummariesByAppGUID[app.GUID].Sort()

		summary := ApplicationSummary{
			Application:      app,
			ProcessSummaries: processSummariesByAppGUID[app.GUID],
			Routes:           routesByAppGUID[app.GUID],
		}

		allSummaries = append(allSummaries, summary)
	}

	return allSummaries, allWarnings, nil
}

func (actor Actor) GetDetailedAppSummary(appName, spaceGUID string, withObfuscatedValues bool) (DetailedApplicationSummary, Warnings, error) {
	var allWarnings Warnings

	app, actorWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	allWarnings = append(allWarnings, actorWarnings...)
	if err != nil {
		return DetailedApplicationSummary{}, actorWarnings, err
	}

	summary, warnings, err := actor.createSummary(app, withObfuscatedValues)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return DetailedApplicationSummary{}, allWarnings, err
	}

	detailedSummary, warnings, err := actor.addDroplet(summary)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return DetailedApplicationSummary{}, allWarnings, err
	}

	detailedSummary, warnings, err = actor.addDeployment(detailedSummary)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return DetailedApplicationSummary{}, allWarnings, err
	}

	return detailedSummary, allWarnings, err
}

func (actor Actor) getProcessSummariesForApps(apps []resources.Application) (map[string]ProcessSummaries, Warnings, error) {
	processSummariesByAppGUID := make(map[string]ProcessSummaries)
	var allWarnings Warnings
	var processes []resources.Process

	warnings, err := batcher.RequestByGUID(toAppGUIDs(apps), func(guids []string) (ccv3.Warnings, error) {
		batch, warnings, err := actor.CloudControllerClient.GetProcesses(ccv3.Query{
			Key: ccv3.AppGUIDFilter, Values: guids,
		})
		processes = append(processes, batch...)
		return warnings, err
	})
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	for _, process := range processes {
		instances, warnings, err := actor.CloudControllerClient.GetProcessInstances(process.GUID)
		allWarnings = append(allWarnings, Warnings(warnings)...)

		if err != nil {
			switch err.(type) {
			case ccerror.ProcessNotFoundError, ccerror.InstanceNotFoundError:
				continue
			default:
				return nil, allWarnings, err
			}
		}

		var instanceDetails []ProcessInstance
		for _, instance := range instances {
			instanceDetails = append(instanceDetails, ProcessInstance(instance))
		}

		processSummary := ProcessSummary{
			Process:         resources.Process(process),
			InstanceDetails: instanceDetails,
		}

		processSummariesByAppGUID[process.AppGUID] = append(processSummariesByAppGUID[process.AppGUID], processSummary)
	}
	return processSummariesByAppGUID, allWarnings, nil
}

func (actor Actor) createSummary(app resources.Application, withObfuscatedValues bool) (ApplicationSummary, Warnings, error) {
	var allWarnings Warnings

	processSummaries, processWarnings, err := actor.getProcessSummariesForApp(app.GUID, withObfuscatedValues)
	allWarnings = append(allWarnings, processWarnings...)
	if err != nil {
		return ApplicationSummary{}, allWarnings, err
	}

	routes, warnings, err := actor.GetApplicationRoutes(app.GUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return ApplicationSummary{}, allWarnings, err
	}

	return ApplicationSummary{
		Application:      app,
		ProcessSummaries: processSummaries,
		Routes:           routes,
	}, allWarnings, nil
}

func (actor Actor) addDroplet(summary ApplicationSummary) (DetailedApplicationSummary, Warnings, error) {
	var allWarnings Warnings

	droplet, warnings, err := actor.GetCurrentDropletByApplication(summary.GUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		if _, ok := err.(actionerror.DropletNotFoundError); !ok {
			return DetailedApplicationSummary{}, allWarnings, err
		}
	}
	return DetailedApplicationSummary{
		ApplicationSummary: summary,
		CurrentDroplet:     droplet,
	}, allWarnings, nil
}

func (actor Actor) addDeployment(detailedSummary DetailedApplicationSummary) (DetailedApplicationSummary, Warnings, error) {
	var allWarnings Warnings

	deployment, warnings, err := actor.GetLatestActiveDeploymentForApp(detailedSummary.GUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil && !errors.Is(err, actionerror.ActiveDeploymentNotFoundError{}) {
		return DetailedApplicationSummary{}, allWarnings, err
	}

	detailedSummary.Deployment = deployment
	return detailedSummary, allWarnings, nil
}

func toAppGUIDs(apps []resources.Application) []string {
	guids := make([]string, len(apps))

	for i, app := range apps {
		guids[i] = app.GUID
	}

	return guids
}
