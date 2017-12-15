package ccv3

import (
	"encoding/json"
	"strconv"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

type ProcessInstance struct {
	Index       int
	State       constant.ProcessInstanceState
	Uptime      int
	CPU         float64
	MemoryUsage uint64
	MemoryQuota uint64
	DiskUsage   uint64
	DiskQuota   uint64
}

// UnmarshalJSON helps unmarshal a V3 Cloud Controller Instance response.
func (instance *ProcessInstance) UnmarshalJSON(data []byte) error {
	var inputInstance struct {
		State string `json:"state"`
		Usage struct {
			CPU  float64 `json:"cpu"`
			Mem  uint64  `json:"mem"`
			Disk uint64  `json:"disk"`
		} `json:"usage"`
		MemQuota  uint64 `json:"mem_quota"`
		DiskQuota uint64 `json:"disk_quota"`
		Index     int    `json:"index"`
		Uptime    int    `json:"uptime"`
	}
	if err := json.Unmarshal(data, &inputInstance); err != nil {
		return err
	}

	instance.State = constant.ProcessInstanceState(inputInstance.State)
	instance.CPU = inputInstance.Usage.CPU
	instance.MemoryUsage = inputInstance.Usage.Mem
	instance.DiskUsage = inputInstance.Usage.Disk

	instance.MemoryQuota = inputInstance.MemQuota
	instance.DiskQuota = inputInstance.DiskQuota
	instance.Index = inputInstance.Index
	instance.Uptime = inputInstance.Uptime

	return nil
}

// DeleteApplicationProcessInstance deletes/stops a particular application's
// process instance.
func (client *Client) DeleteApplicationProcessInstance(appGUID string, processType string, instanceIndex int) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteApplicationProcessInstanceRequest,
		URIParams: map[string]string{
			"app_guid": appGUID,
			"type":     processType,
			"index":    strconv.Itoa(instanceIndex),
		},
	})
	if err != nil {
		return nil, err
	}

	var response cloudcontroller.Response
	err = client.connection.Make(request, &response)

	return response.Warnings, err
}

// GetProcessInstances lists instance stats for a given process.
func (client *Client) GetProcessInstances(processGUID string) ([]ProcessInstance, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetProcessInstancesRequest,
		URIParams:   map[string]string{"process_guid": processGUID},
	})
	if err != nil {
		return nil, nil, err
	}

	var fullInstancesList []ProcessInstance
	warnings, err := client.paginate(request, ProcessInstance{}, func(item interface{}) error {
		if instance, ok := item.(ProcessInstance); ok {
			fullInstancesList = append(fullInstancesList, instance)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   ProcessInstance{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullInstancesList, warnings, err
}
