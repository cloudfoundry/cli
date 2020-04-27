package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
)

type ApplicationSummary struct {
	Application
	ProcessSummaries ProcessSummaries
	Routes           []resources.Route
}

// v7action.DetailedApplicationSummary represents an application with its processes and droplet.
type DetailedApplicationSummary struct {
	ApplicationSummary
	CurrentDroplet Droplet
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

func (actor Actor) GetAppSummariesForSpace(spaceGUID string, labelSelector string) ([]ApplicationSummary, Warnings, error) {
	var allWarnings Warnings
	var allSummaries []ApplicationSummary

	keys := []ccv3.Query{
		{Key: ccv3.SpaceGUIDFilter, Values: []string{spaceGUID}},
		{Key: ccv3.OrderBy, Values: []string{ccv3.NameOrder}},
	}
	if len(labelSelector) > 0 {
		keys = append(keys, ccv3.Query{Key: ccv3.LabelSelectorFilter, Values: []string{labelSelector}})
	}
	apps, warnings, err := actor.CloudControllerClient.GetApplications(keys...)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	processes, warnings, err := actor.CloudControllerClient.GetProcesses(ccv3.Query{
		Key: ccv3.AppGUIDFilter, Values: toAppGUIDs(apps),
	})
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	processSummariesByAppGUID := make(map[string]ProcessSummaries, len(apps))
	for _, process := range processes {
		instances, warnings, err := actor.CloudControllerClient.GetProcessInstances(process.GUID)
		allWarnings = append(allWarnings, Warnings(warnings)...)
		if err != nil {
			return nil, allWarnings, err
		}

		var instanceDetails []ProcessInstance
		for _, instance := range instances {
			instanceDetails = append(instanceDetails, ProcessInstance(instance))
		}

		processSummary := ProcessSummary{
			Process:         Process(process),
			InstanceDetails: instanceDetails,
		}

		processSummariesByAppGUID[process.AppGUID] = append(processSummariesByAppGUID[process.AppGUID], processSummary)
	}

	spaceRoutes, warnings, err := actor.CloudControllerClient.GetRoutes(ccv3.Query{
		Key:    ccv3.SpaceGUIDFilter,
		Values: []string{spaceGUID},
	})
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	routesByAppGUID := make(map[string][]resources.Route)
	for _, route := range spaceRoutes {
		for _, dest := range route.Destinations {
			routesByAppGUID[dest.App.GUID] = append(routesByAppGUID[dest.App.GUID], route)
		}
	}

	for _, app := range apps {
		//routes, warnings, err := actor.CloudControllerClient.GetApplicationRoutes(app.GUID)
		//allWarnings = append(allWarnings, warnings...)
		//if err != nil {
		//	return nil, allWarnings, err
		//}

		processSummariesByAppGUID[app.GUID].Sort()

		summary := ApplicationSummary{
			Application:      actor.convertCCToActorApplication(app),
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

	return detailedSummary, allWarnings, err
}

func (actor Actor) createSummary(app Application, withObfuscatedValues bool) (ApplicationSummary, Warnings, error) {
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
		Application: Application{
			Name:                app.Name,
			GUID:                app.GUID,
			State:               app.State,
			LifecycleType:       app.LifecycleType,
			LifecycleBuildpacks: app.LifecycleBuildpacks,
		},
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

func toAppGUIDs(apps []ccv3.Application) []string {
	guids := make([]string, len(apps))

	for i, app := range apps {
		guids[i] = app.GUID
	}

	return guids
}
