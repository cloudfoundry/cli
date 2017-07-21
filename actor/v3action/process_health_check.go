package v3action

type ProcessHealthCheck struct {
	ProcessType     string
	HealthCheckType string
	Endpoint        string
}

func (actor Actor) GetApplicationProcessHealthChecksByNameAndSpace(appName string, spaceGUID string) ([]ProcessHealthCheck, Warnings, error) {
	app, allWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return nil, allWarnings, err
	}

	ccv3Processes, warnings, err := actor.CloudControllerClient.GetApplicationProcesses(app.GUID)
	allWarnings = append(allWarnings, Warnings(warnings)...)
	if err != nil {
		return nil, allWarnings, err
	}

	var processHealthChecks []ProcessHealthCheck
	for _, ccv3Process := range ccv3Processes {
		processHealthCheck := ProcessHealthCheck{
			ProcessType:     ccv3Process.Type,
			HealthCheckType: ccv3Process.HealthCheck.Type,
			Endpoint:        ccv3Process.HealthCheck.Data.Endpoint,
		}
		processHealthChecks = append(processHealthChecks, processHealthCheck)
	}

	return processHealthChecks, allWarnings, nil
}
