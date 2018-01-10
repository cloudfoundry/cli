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
	GUID                string           `json:"guid"`
	Type                string           `json:"type"`
	HealthCheckType     string           `json:"-"`
	HealthCheckEndpoint string           `json:"-"`
	Instances           types.NullInt    `json:"instances"`
	MemoryInMB          types.NullUint64 `json:"memory_in_mb"`
	DiskInMB            types.NullUint64 `json:"disk_in_mb"`
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

	ccProcess.HealthCheck.Type = p.HealthCheckType
	if p.HealthCheckEndpoint != "" {
		ccProcess.HealthCheck.Data.Endpoint = p.HealthCheckEndpoint
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
				Endpoint string `json:"endpoint"`
			} `json:"data"`
		} `json:"health_check"`
	}

	ccProcess.rawProcess = (*rawProcess)(p)
	if err := json.Unmarshal(data, &ccProcess); err != nil {
		return err
	}

	p.HealthCheckEndpoint = ccProcess.HealthCheck.Data.Endpoint
	p.HealthCheckType = ccProcess.HealthCheck.Type
	return nil
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
func (client *Client) PatchApplicationProcessHealthCheck(processGUID string, processHealthCheckType string, processHealthCheckEndpoint string) (Process, Warnings, error) {
	body, err := json.Marshal(Process{
		HealthCheckType:     processHealthCheckType,
		HealthCheckEndpoint: processHealthCheckEndpoint,
	})
	if err != nil {
		return Process{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PatchApplicationProcessHealthCheckRequest,
		Body:        bytes.NewReader(body),
		URIParams:   internal.Params{"process_guid": processGUID},
	})
	if err != nil {
		return Process{}, nil, err
	}

	var responceProcess Process
	response := cloudcontroller.Response{
		Result: &responceProcess,
	}
	err = client.connection.Make(request, &response)
	return responceProcess, response.Warnings, err
}

// CreateApplicationProcessScale updates process instances count, memory or disk
func (client *Client) CreateApplicationProcessScale(appGUID string, process Process) (Process, Warnings, error) {
	ccProcessScale := struct {
		Instances  json.Number `json:"instances,omitempty"`
		MemoryInMB json.Number `json:"memory_in_mb,omitempty"`
		DiskInMB   json.Number `json:"disk_in_mb,omitempty"`
	}{}

	if process.Instances.IsSet {
		ccProcessScale.Instances = json.Number(fmt.Sprint(process.Instances.Value))
	}
	if process.MemoryInMB.IsSet {
		ccProcessScale.MemoryInMB = json.Number(fmt.Sprint(process.MemoryInMB.Value))
	}
	if process.DiskInMB.IsSet {
		ccProcessScale.DiskInMB = json.Number(fmt.Sprint(process.DiskInMB.Value))
	}

	body, err := json.Marshal(ccProcessScale)
	if err != nil {
		return Process{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostApplicationProcessScaleRequest,
		Body:        bytes.NewReader(body),
		URIParams:   internal.Params{"app_guid": appGUID, "type": process.Type},
	})
	if err != nil {
		return Process{}, nil, err
	}

	var responceProcess Process
	response := cloudcontroller.Response{
		Result: &responceProcess,
	}
	err = client.connection.Make(request, &response)
	return responceProcess, response.Warnings, err
}
