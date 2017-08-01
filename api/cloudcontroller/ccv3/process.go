package ccv3

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
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

	request, _ := client.newHTTPRequest(requestOptions{
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
