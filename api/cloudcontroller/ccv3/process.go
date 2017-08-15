package ccv3

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/types"
)

type Process struct {
	GUID        string             `json:"guid"`
	Type        string             `json:"type"`
	HealthCheck ProcessHealthCheck `json:"health_check"`
	Instances   int                `json:"instances"`
	MemoryInMB  int                `json:"memory_in_mb"`
	DiskInMB    int                `json:"disk_in_mb"`
}

type ProcessHealthCheck struct {
	Type string                 `json:"type"`
	Data ProcessHealthCheckData `json:"data"`
}

type ProcessHealthCheckData struct {
	Endpoint string `json:"endpoint"`
}

type ProcessScaleOptions struct {
	Instances  types.NullInt    `json:"instances"`
	MemoryInMB types.NullUint64 `json:"memory_in_mb"`
	DiskInMB   types.NullUint64 `json:"disk_in_mb"`
}

func (p Process) MarshalJSON() ([]byte, error) {
	var ccProcess struct {
		HealthCheck struct {
			Type string `json:"type"`
			Data struct {
				Endpoint interface{} `json:"endpoint"`
			} `json:"data"`
		} `json:"health_check"`
	}

	ccProcess.HealthCheck.Type = p.HealthCheck.Type
	if p.HealthCheck.Data.Endpoint != "" {
		ccProcess.HealthCheck.Data.Endpoint = p.HealthCheck.Data.Endpoint
	}
	return json.Marshal(ccProcess)
}

// GetApplicationProcesses lists processes for a given app
func (client *Client) GetApplicationProcesses(appGUID string) ([]Process, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetAppProcessesRequest,
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

// GetApplicationProcessByType returns application process of specified type
func (client *Client) GetApplicationProcessByType(appGUID string, processType string) (Process, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetApplicationProcessByTypeRequest,
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

// PatchApplicationProcessHealthCheck updates application health check type
func (client *Client) PatchApplicationProcessHealthCheck(processGUID string, processHealthCheckType string, processHealthCheckEndpoint string) (Warnings, error) {
	body, err := json.Marshal(Process{
		HealthCheck: ProcessHealthCheck{
			Type: processHealthCheckType,
			Data: ProcessHealthCheckData{
				Endpoint: processHealthCheckEndpoint,
			}}})
	if err != nil {
		return nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PatchApplicationProcessHealthCheckRequest,
		Body:        bytes.NewReader(body),
		URIParams:   internal.Params{"process_guid": processGUID},
	})
	if err != nil {
		return nil, err
	}

	var response cloudcontroller.Response
	err = client.connection.Make(request, &response)
	return response.Warnings, err
}

func (client *Client) CreateApplicationProcessScale(appGUID string, processType string, scaleOptions ProcessScaleOptions) (Process, Warnings, error) {
	ccProcessScale := struct {
		Instances  int    `json:"instances,omitempty"`
		MemoryInMB uint64 `json:"memory_in_mb,omitempty"`
		DiskInMB   uint64 `json:"disk_in_mb,omitempty"`
	}{}

	if scaleOptions.Instances.IsSet {
		ccProcessScale.Instances = scaleOptions.Instances.Value
	}
	if scaleOptions.MemoryInMB.IsSet {
		ccProcessScale.MemoryInMB = scaleOptions.MemoryInMB.Value
	}
	if scaleOptions.DiskInMB.IsSet {
		ccProcessScale.DiskInMB = scaleOptions.DiskInMB.Value
	}

	body, err := json.Marshal(ccProcessScale)
	if err != nil {
		return Process{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostApplicationProcessScaleRequest,
		Body:        bytes.NewReader(body),
		URIParams:   internal.Params{"app_guid": appGUID, "type": processType},
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
