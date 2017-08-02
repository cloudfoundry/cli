package v3action

import "net/url"

// ApplicationSummary represents an application with its processes and droplet.
type ApplicationSummary struct {
	Application
	Processes      Processes
	CurrentDroplet Droplet
}

// GetApplicationSummaryByNameAndSpace returns an application with process and
// instance stats.
func (actor Actor) GetApplicationSummaryByNameAndSpace(appName string,
	spaceGUID string) (ApplicationSummary, Warnings, error) {
	app, allWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return ApplicationSummary{}, allWarnings, err
	}

	processes, processWarnings, err := actor.getProcessesForApp(app.GUID)
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
		for _, ccv3Buildpack := range ccv3Droplets[0].Buildpacks {
			droplet.Buildpacks = append(droplet.Buildpacks, Buildpack(ccv3Buildpack))
		}
	}

	summary := ApplicationSummary{
		Application:    app,
		Processes:      processes,
		CurrentDroplet: droplet,
	}
	return summary, allWarnings, nil
}

func (actor Actor) GetInstancesByApplicationAndProcessType(appGUID string, processType string) (Process, Warnings, error) {
	var allWarnings Warnings

	ccv3Process, warnings, err := actor.CloudControllerClient.GetApplicationProcessByType(appGUID, processType)
	allWarnings = Warnings(warnings)
	if err != nil {
		return Process{}, allWarnings, err
	}

	processGUID := ccv3Process.GUID
	instances, warnings, err := actor.CloudControllerClient.GetProcessInstances(processGUID)
	allWarnings = append(allWarnings, Warnings(warnings)...)
	if err != nil {
		return Process{}, allWarnings, err
	}

	process := Process{
		Type:       ccv3Process.Type,
		MemoryInMB: ccv3Process.MemoryInMB,
		DiskInMB:   ccv3Process.DiskInMB,
	}
	for _, instance := range instances {
		process.Instances = append(process.Instances, Instance(instance))
	}

	return process, allWarnings, nil
}

func (actor Actor) getProcessesForApp(appGUID string) (Processes, Warnings, error) {
	var allWarnings Warnings

	ccv3Processes, warnings, err := actor.CloudControllerClient.GetApplicationProcesses(appGUID)
	allWarnings = Warnings(warnings)
	if err != nil {
		return nil, allWarnings, err
	}

	var processes Processes
	for _, ccv3Process := range ccv3Processes {
		processGUID := ccv3Process.GUID
		instances, warnings, err := actor.CloudControllerClient.GetProcessInstances(processGUID)
		allWarnings = append(allWarnings, Warnings(warnings)...)
		if err != nil {
			return nil, allWarnings, err
		}

		process := Process{
			Type:       ccv3Process.Type,
			Instances:  []Instance{},
			MemoryInMB: ccv3Process.MemoryInMB,
		}
		for _, instance := range instances {
			process.Instances = append(process.Instances, Instance(instance))
		}

		processes = append(processes, process)
	}

	return processes, allWarnings, nil
}
