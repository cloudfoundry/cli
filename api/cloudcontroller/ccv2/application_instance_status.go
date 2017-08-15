package ccv2

import (
	"encoding/json"
	"strconv"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// ApplicationInstanceStatus represents a Cloud Controller Application Instance.
type ApplicationInstanceStatus struct {
	// CPU is the instance's CPU utilization percentage.
	CPU float64

	// Disk is the instance's disk usage in bytes.
	Disk int

	// DiskQuota is the instance's allowed disk usage in bytes.
	DiskQuota int

	// ID is the instance ID.
	ID int

	// IsolationSegment that the app is currently running on.
	IsolationSegment string

	// Memory is the instance's memory usage in bytes.
	Memory int

	// MemoryQuota is the instance's allowed memory usage in bytes.
	MemoryQuota int

	// State is the instance's state.
	State ApplicationInstanceState

	// Uptime is the number of seconds the instance has been running.
	Uptime int
}

// UnmarshalJSON helps unmarshal a Cloud Controller application instance
// response.
func (instance *ApplicationInstanceStatus) UnmarshalJSON(data []byte) error {
	var ccInstance struct {
		State            string `json:"state"`
		IsolationSegment string `json:"isolation_segment"`
		Stats            struct {
			Usage struct {
				Disk   int     `json:"disk"`
				Memory int     `json:"mem"`
				CPU    float64 `json:"cpu"`
			} `json:"usage"`
			MemoryQuota int `json:"mem_quota"`
			DiskQuota   int `json:"disk_quota"`
			Uptime      int `json:"uptime"`
		} `json:"stats"`
	}
	if err := json.Unmarshal(data, &ccInstance); err != nil {
		return err
	}

	instance.CPU = ccInstance.Stats.Usage.CPU
	instance.Disk = ccInstance.Stats.Usage.Disk
	instance.DiskQuota = ccInstance.Stats.DiskQuota
	instance.IsolationSegment = ccInstance.IsolationSegment
	instance.Memory = ccInstance.Stats.Usage.Memory
	instance.MemoryQuota = ccInstance.Stats.MemoryQuota
	instance.State = ApplicationInstanceState(ccInstance.State)
	instance.Uptime = ccInstance.Stats.Uptime

	return nil
}

// GetApplicationInstanceStatusesByApplication returns a list of
// ApplicationInstance for a given application. Given the state of an
// application, it might skip some application instances.
func (client *Client) GetApplicationInstanceStatusesByApplication(guid string) (map[int]ApplicationInstanceStatus, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetAppStatsRequest,
		URIParams:   Params{"app_guid": guid},
	})
	if err != nil {
		return nil, nil, err
	}

	var instances map[string]ApplicationInstanceStatus
	response := cloudcontroller.Response{
		Result: &instances,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return nil, response.Warnings, err
	}

	returnedInstances := map[int]ApplicationInstanceStatus{}
	for instanceID, instance := range instances {
		id, convertErr := strconv.Atoi(instanceID)
		if convertErr != nil {
			return nil, response.Warnings, convertErr
		}
		instance.ID = id
		returnedInstances[id] = instance
	}

	return returnedInstances, response.Warnings, nil
}
