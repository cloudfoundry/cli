package ccv3

import (
	"bytes"
	"encoding/json"
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/types"
)

type Process struct {
	GUID string
	Type string
	// Command is the process start command. Note: This value will be obfuscated when obtained from listing.
	Command                      string
	HealthCheckType              string
	HealthCheckEndpoint          string
	HealthCheckInvocationTimeout int
	Instances                    types.NullInt
	MemoryInMB                   types.NullUint64
	DiskInMB                     types.NullUint64
}

func (p Process) MarshalJSON() ([]byte, error) {
	type healthCheck struct {
		Type string `json:"type"`
		Data struct {
			Endpoint          interface{} `json:"endpoint"`
			InvocationTimeout int         `json:"invocation_timeout,omitempty"`
		} `json:"data"`
	}

	var ccProcess struct {
		Command    string      `json:"command,omitempty"`
		Instances  json.Number `json:"instances,omitempty"`
		MemoryInMB json.Number `json:"memory_in_mb,omitempty"`
		DiskInMB   json.Number `json:"disk_in_mb,omitempty"`

		HealthCheck *healthCheck `json:"health_check,omitempty"`
	}

	ccProcess.Command = p.Command
	if p.Instances.IsSet {
		ccProcess.Instances = json.Number(fmt.Sprint(p.Instances.Value))
	}
	if p.MemoryInMB.IsSet {
		ccProcess.MemoryInMB = json.Number(fmt.Sprint(p.MemoryInMB.Value))
	}
	if p.DiskInMB.IsSet {
		ccProcess.DiskInMB = json.Number(fmt.Sprint(p.DiskInMB.Value))
	}

	if p.HealthCheckType != "" || p.HealthCheckEndpoint != "" || p.HealthCheckInvocationTimeout != 0 {
		ccProcess.HealthCheck = new(healthCheck)
		ccProcess.HealthCheck.Type = p.HealthCheckType
		ccProcess.HealthCheck.Data.InvocationTimeout = p.HealthCheckInvocationTimeout
		if p.HealthCheckEndpoint != "" {
			ccProcess.HealthCheck.Data.Endpoint = p.HealthCheckEndpoint
		}
	}

	return json.Marshal(ccProcess)
}

func (p *Process) UnmarshalJSON(data []byte) error {
	var ccProcess struct {
		Command    string           `json:"command"`
		DiskInMB   types.NullUint64 `json:"disk_in_mb"`
		GUID       string           `json:"guid"`
		Instances  types.NullInt    `json:"instances"`
		MemoryInMB types.NullUint64 `json:"memory_in_mb"`
		Type       string           `json:"type"`

		HealthCheck struct {
			Type string `json:"type"`
			Data struct {
				Endpoint          string `json:"endpoint"`
				InvocationTimeout int    `json:"invocation_timeout"`
			} `json:"data"`
		} `json:"health_check"`
	}

	err := cloudcontroller.DecodeJSON(data, &ccProcess)
	if err != nil {
		return err
	}

	p.Command = ccProcess.Command
	p.DiskInMB = ccProcess.DiskInMB
	p.GUID = ccProcess.GUID
	p.HealthCheckEndpoint = ccProcess.HealthCheck.Data.Endpoint
	p.HealthCheckInvocationTimeout = ccProcess.HealthCheck.Data.InvocationTimeout
	p.HealthCheckType = ccProcess.HealthCheck.Type
	p.Instances = ccProcess.Instances
	p.MemoryInMB = ccProcess.MemoryInMB
	p.Type = ccProcess.Type

	return nil
}

// CreateApplicationProcessScale updates process instances count, memory or disk
func (client *Client) CreateApplicationProcessScale(appGUID string, process Process) (Process, Warnings, error) {
	body, err := json.Marshal(process)
	if err != nil {
		return Process{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostApplicationProcessActionScaleRequest,
		Body:        bytes.NewReader(body),
		URIParams:   internal.Params{"app_guid": appGUID, "type": process.Type},
	})
	if err != nil {
		return Process{}, nil, err
	}

	var responseProcess Process
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &responseProcess,
	}
	err = client.connection.Make(request, &response)
	return responseProcess, response.Warnings, err
}

// GetApplicationProcessByType returns application process of specified type
func (client *Client) GetApplicationProcessByType(appGUID string, processType string) (Process, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetApplicationProcessRequest,
		URIParams: map[string]string{
			"app_guid": appGUID,
			"type":     processType,
		},
	})
	if err != nil {
		return Process{}, nil, err
	}
	var process Process
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &process,
	}

	err = client.connection.Make(request, &response)
	return process, response.Warnings, err
}

// GetApplicationProcesses lists processes for a given application. **Note**:
// Due to security, the API obfuscates certain values such as `command`.
func (client *Client) GetApplicationProcesses(appGUID string) ([]Process, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetApplicationProcessesRequest,
		URIParams:   map[string]string{"app_guid": appGUID},
	})
	if err != nil {
		return nil, nil, err
	}

	var fullProcessesList []Process
	warnings, err := client.paginate(request, Process{}, func(item interface{}) error {
		if process, ok := item.(Process); ok {
			fullProcessesList = append(fullProcessesList, process)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Process{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullProcessesList, warnings, err
}

// UpdateProcess updates the process's health check settings. GUID is always
// required; HealthCheckType is only required when updating health check
// settings.
func (client *Client) UpdateProcess(process Process) (Process, Warnings, error) {
	body, err := json.Marshal(Process{
		Command:                      process.Command,
		HealthCheckType:              process.HealthCheckType,
		HealthCheckEndpoint:          process.HealthCheckEndpoint,
		HealthCheckInvocationTimeout: process.HealthCheckInvocationTimeout,
	})
	if err != nil {
		return Process{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PatchProcessRequest,
		Body:        bytes.NewReader(body),
		URIParams:   internal.Params{"process_guid": process.GUID},
	})
	if err != nil {
		return Process{}, nil, err
	}

	var responseProcess Process
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &responseProcess,
	}
	err = client.connection.Make(request, &response)
	return responseProcess, response.Warnings, err
}
