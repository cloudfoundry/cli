package ccv3

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

type Instance struct {
	State       string
	CPU         float64
	MemoryUsage uint64
	DiskUsage   uint64
	MemoryQuota uint64
	DiskQuota   uint64
	Index       int
	Uptime      int
}

// UnmarshalJSON helps unmarshal a V3 Cloud Controller Instance response.
func (instance *Instance) UnmarshalJSON(data []byte) error {
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

	instance.State = inputInstance.State
	instance.CPU = inputInstance.Usage.CPU
	instance.MemoryUsage = inputInstance.Usage.Mem
	instance.DiskUsage = inputInstance.Usage.Disk

	instance.MemoryQuota = inputInstance.MemQuota
	instance.DiskQuota = inputInstance.DiskQuota
	instance.Index = inputInstance.Index
	instance.Uptime = inputInstance.Uptime

	return nil
}

// GetProcessInstances lists instance stats for a given process.
func (client *Client) GetProcessInstances(processGUID string) ([]Instance, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetProcessInstancesRequest,
		URIParams:   map[string]string{"guid": processGUID},
	})
	if err != nil {
		return nil, nil, err
	}

	var fullInstancesList []Instance
	warnings, err := client.paginate(request, Instance{}, func(item interface{}) error {
		if instance, ok := item.(Instance); ok {
			fullInstancesList = append(fullInstancesList, instance)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Instance{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullInstancesList, warnings, err
}
