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
	GUID                         string           `json:"guid"`
	Type                         string           `json:"type"`
	HealthCheckType              string           `json:"-"`
	HealthCheckEndpoint          string           `json:"-"`
	HealthCheckInvocationTimeout int              `json:"-"`
	Instances                    types.NullInt    `json:"instances,omitempty"`
	MemoryInMB                   types.NullUint64 `json:"memory_in_mb,omitempty"`
	DiskInMB                     types.NullUint64 `json:"disk_in_mb,omitempty"`
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
		Instances  json.Number `json:"instances,omitempty"`
		MemoryInMB json.Number `json:"memory_in_mb,omitempty"`
		DiskInMB   json.Number `json:"disk_in_mb,omitempty"`

		HealthCheck *healthCheck `json:"health_check,omitempty"`
	}

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
	type rawProcess Process
	var ccProcess struct {
		*rawProcess

		HealthCheck struct {
			Type string `json:"type"`
			Data struct {
				Endpoint          string `json:"endpoint"`
				InvocationTimeout int    `json:"invocation_timeout"`
			} `json:"data"`
		} `json:"health_check"`
	}

	ccProcess.rawProcess = (*rawProcess)(p)
	err := cloudcontroller.DecodeJSON(data, &ccProcess)
	if err != nil {
		return err
	}

	p.HealthCheckEndpoint = ccProcess.HealthCheck.Data.Endpoint
	p.HealthCheckType = ccProcess.HealthCheck.Type
	p.HealthCheckInvocationTimeout = ccProcess.HealthCheck.Data.InvocationTimeout

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
		Result: &responseProcess,
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
		Result: &process,
	}

	err = client.connection.Make(request, &response)
	return process, response.Warnings, err
}

// GetApplicationProcesses lists processes for a given app
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

// PatchApplicationProcessHealthCheck updates application health check type
func (client *Client) PatchApplicationProcessHealthCheck(processGUID string, processHealthCheckType string, processHealthCheckEndpoint string, processHealthCheckInvocationTimeout int) (Process, Warnings, error) {
	body, err := json.Marshal(Process{
		HealthCheckType:              processHealthCheckType,
		HealthCheckEndpoint:          processHealthCheckEndpoint,
		HealthCheckInvocationTimeout: processHealthCheckInvocationTimeout,
	})
	if err != nil {
		return Process{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PatchProcessRequest,
		Body:        bytes.NewReader(body),
		URIParams:   internal.Params{"process_guid": processGUID},
	})
	if err != nil {
		return Process{}, nil, err
	}

	var responseProcess Process
	response := cloudcontroller.Response{
		Result: &responseProcess,
	}
	err = client.connection.Make(request, &response)
	return responseProcess, response.Warnings, err
}
