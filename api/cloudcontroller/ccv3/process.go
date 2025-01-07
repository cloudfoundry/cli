package ccv3

import (
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/v8/resources"
)

// CreateApplicationProcessScale updates process instances count, memory or disk
func (client *Client) CreateApplicationProcessScale(appGUID string, process resources.Process) (resources.Process, Warnings, error) {
	var responseBody resources.Process

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PostApplicationProcessActionScaleRequest,
		URIParams:    internal.Params{"app_guid": appGUID, "type": process.Type},
		RequestBody:  process,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// GetApplicationProcessByType returns application process of specified type
func (client *Client) GetApplicationProcessByType(appGUID string, processType string) (resources.Process, Warnings, error) {
	var responseBody resources.Process

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetApplicationProcessRequest,
		URIParams:    internal.Params{"app_guid": appGUID, "type": processType},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// GetApplicationProcesses lists processes for a given application. **Note**:
// Due to security, the API obfuscates certain values such as `command`.
func (client *Client) GetApplicationProcesses(appGUID string) ([]resources.Process, Warnings, error) {
	var processes []resources.Process

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetApplicationProcessesRequest,
		URIParams:    internal.Params{"app_guid": appGUID},
		ResponseBody: resources.Process{},
		AppendToList: func(item interface{}) error {
			processes = append(processes, item.(resources.Process))
			return nil
		},
	})

	return processes, warnings, err
}

// GetNewApplicationProcesses gets processes for an application in the middle of a deployment.
// The app's processes will include a web process that will be removed when the deployment completes,
// so exclude that soon-to-be-removed process from the result.
func (client *Client) GetNewApplicationProcesses(appGUID string, deploymentGUID string) ([]resources.Process, Warnings, error) {
	var allWarnings Warnings

	deployment, warnings, err := client.GetDeployment(deploymentGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	allProcesses, warnings, err := client.GetApplicationProcesses(appGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	var newWebProcessGUID string
	for _, process := range deployment.NewProcesses {
		if process.Type == constant.ProcessTypeWeb {
			newWebProcessGUID = process.GUID
		}
	}

	var processesList []resources.Process
	for _, process := range allProcesses {
		if process.Type == constant.ProcessTypeWeb {
			if process.GUID == newWebProcessGUID {
				processesList = append(processesList, process)
			}
		} else {
			processesList = append(processesList, process)
		}
	}

	return processesList, allWarnings, nil
}

// GetProcess returns a process with the given guid
func (client *Client) GetProcess(processGUID string) (resources.Process, Warnings, error) {
	var responseBody resources.Process

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetProcessRequest,
		URIParams:    internal.Params{"process_guid": processGUID},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

func (client Client) GetProcesses(query ...Query) ([]resources.Process, Warnings, error) {
	var processes []resources.Process

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetProcessesRequest,
		Query:        query,
		ResponseBody: resources.Process{},
		AppendToList: func(item interface{}) error {
			processes = append(processes, item.(resources.Process))
			return nil
		},
	})

	return processes, warnings, err
}

// UpdateProcess updates the process's command or health check settings. GUID
// is always required; HealthCheckType is only required when updating health
// check settings.
func (client *Client) UpdateProcess(process resources.Process) (resources.Process, Warnings, error) {
	var responseBody resources.Process

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName: internal.PatchProcessRequest,
		URIParams:   internal.Params{"process_guid": process.GUID},
		RequestBody: resources.Process{
			Command:                      process.Command,
			HealthCheckType:              process.HealthCheckType,
			HealthCheckEndpoint:          process.HealthCheckEndpoint,
			HealthCheckTimeout:           process.HealthCheckTimeout,
			HealthCheckInvocationTimeout: process.HealthCheckInvocationTimeout,
		},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}
