package ccv2

import (
	"encoding/json"
	"sort"
	"strconv"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// ApplicationInstanceStatusState reflects the state of the individual app
// instance.
type ApplicationInstanceStatusState string

const (
	ApplicationInstanceCrashed  ApplicationInstanceStatusState = "CRASHED"
	ApplicationInstanceDown                                    = "DOWN"
	ApplicationInstanceRunning                                 = "RUNNING"
	ApplicationInstanceStarting                                = "STARTING"
	ApplicationInstanceUnknown                                 = "UNKNOWN"
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

	// Memory is the instance's memory usage in bytes.
	Memory int

	// MemoryQuota is the instance's allowed memory usage in bytes.
	MemoryQuota int

	// State is the instance's state.
	State ApplicationInstanceStatusState

	// Uptime is the number of seconds the instance has been running.
	Uptime int
}

// UnmarshalJSON helps unmarshal a Cloud Controller application instance
// response.
func (instance *ApplicationInstanceStatus) UnmarshalJSON(data []byte) error {
	var ccInstance struct {
		State string `json:"state"`
		Stats struct {
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
	instance.Memory = ccInstance.Stats.Usage.Memory
	instance.MemoryQuota = ccInstance.Stats.MemoryQuota
	instance.State = ApplicationInstanceStatusState(ccInstance.State)
	instance.Uptime = ccInstance.Stats.Uptime

	return nil
}

// GetApplicationInstanceStatusesByApplication returns a list of
// ApplicationInstance for a given application. Given the state of an
// application, it might skip some application instances.
func (client *Client) GetApplicationInstanceStatusesByApplication(guid string) ([]ApplicationInstanceStatus, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.AppInstanceStats,
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

	sortedIDs, err := client.sortedInstanceKeys(instances)
	if err != nil {
		return nil, response.Warnings, err
	}

	var sortedInstances []ApplicationInstanceStatus
	for _, instanceID := range sortedIDs {
		instance := instances[strconv.Itoa(instanceID)]
		instance.ID = instanceID
		sortedInstances = append(sortedInstances, instance)
	}

	return sortedInstances, response.Warnings, err
}

func (client *Client) sortedInstanceKeys(instances map[string]ApplicationInstanceStatus) ([]int, error) {
	var keys []int
	for key, _ := range instances {
		id, err := strconv.Atoi(key)
		if err != nil {
			return nil, err
		}
		keys = append(keys, id)
	}
	sort.Ints(keys)

	return keys, nil
}
