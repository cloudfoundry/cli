package v3action

// ApplicationSummary represents an application with its processes and droplet.
type ApplicationSummary struct {
	Application
	Processes      []Process
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

	ccv3Droplet, warnings, err := actor.CloudControllerClient.GetApplicationCurrentDroplet(app.GUID)
	allWarnings = append(allWarnings, Warnings(warnings)...)
	if err != nil {
		return ApplicationSummary{}, allWarnings, err
	}
	droplet := Droplet{
		Stack: ccv3Droplet.Stack,
	}
	for _, ccv3Buildpack := range ccv3Droplet.Buildpacks {
		droplet.Buildpacks = append(droplet.Buildpacks, Buildpack(ccv3Buildpack))
	}

	ccv3Processes, warnings, err := actor.CloudControllerClient.GetApplicationProcesses(app.GUID)
	allWarnings = append(allWarnings, Warnings(warnings)...)
	if err != nil {
		return ApplicationSummary{}, allWarnings, err
	}

	var processes []Process
	for _, ccv3Process := range ccv3Processes {
		processGUID := ccv3Process.GUID
		instances, warnings, err := actor.CloudControllerClient.GetProcessInstances(processGUID)
		allWarnings = append(allWarnings, Warnings(warnings)...)
		if err != nil {
			return ApplicationSummary{}, allWarnings, err
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

	summary := ApplicationSummary{
		Application:    app,
		Processes:      processes,
		CurrentDroplet: droplet,
	}
	return summary, allWarnings, nil
}
