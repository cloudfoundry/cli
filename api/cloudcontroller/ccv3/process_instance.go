package ccv3

import (
	"fmt"
	"strconv"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/types"
)

// ProcessInstance represents a single process instance for a particular
// application.
type ProcessInstance struct {
	// CPU is the current CPU usage of the instance.
	CPU float64
	// CPU Entitlement is the current CPU entitlement usage of the instance.
	CPUEntitlement types.NullFloat64
	// Details is information about errors placing the instance.
	Details string
	// DiskQuota is the maximum disk the instance is allowed to use.
	DiskQuota uint64
	// DiskUsage is the current disk usage of the instance.
	DiskUsage uint64
	// Index is the index of the instance.
	Index int64
	// Isolation segment is the current isolation segment that the instance is
	// running on. The value is empty when the instance is not placed on a
	// particular isolation segment.
	IsolationSegment string
	// MemoryQuota is the maximum memory the instance is allowed to use.
	MemoryQuota uint64
	// MemoryUsage is the current memory usage of the instance.
	MemoryUsage uint64
	// LogRateLimit is the maximum rate that the instance is allowed to log.
	LogRateLimit int64
	// LogRate is the current rate that the instance is logging.
	LogRate uint64
	// State is the state of the instance.
	State constant.ProcessInstanceState
	// Routeable is the readiness state of the instance, can be true, false or null.
	Routable *bool
	// Type is the process type for the instance.
	Type string
	// Uptime is the duration that the instance has been running.
	Uptime time.Duration
}

// UnmarshalJSON helps unmarshal a V3 Cloud Controller Instance response.
func (instance *ProcessInstance) UnmarshalJSON(data []byte) error {
	var inputInstance struct {
		Details          string `json:"details"`
		DiskQuota        uint64 `json:"disk_quota"`
		Index            int64  `json:"index"`
		IsolationSegment string `json:"isolation_segment"`
		MemQuota         uint64 `json:"mem_quota"`
		LogRateLimit     int64  `json:"log_rate_limit"`
		State            string `json:"state"`
		Routable         *bool  `json:"routable"`
		Type             string `json:"type"`
		Uptime           int64  `json:"uptime"`
		Usage            struct {
			CPU            float64           `json:"cpu"`
			CPUEntitlement types.NullFloat64 `json:"cpu_entitlement"`
			Mem            uint64            `json:"mem"`
			Disk           uint64            `json:"disk"`
			LogRate        uint64            `json:"log_rate"`
		} `json:"usage"`
	}

	err := cloudcontroller.DecodeJSON(data, &inputInstance)
	if err != nil {
		return err
	}

	instance.CPU = inputInstance.Usage.CPU
	instance.CPUEntitlement = inputInstance.Usage.CPUEntitlement
	instance.Details = inputInstance.Details
	instance.DiskQuota = inputInstance.DiskQuota
	instance.DiskUsage = inputInstance.Usage.Disk
	instance.Index = inputInstance.Index
	instance.IsolationSegment = inputInstance.IsolationSegment
	instance.MemoryQuota = inputInstance.MemQuota
	instance.MemoryUsage = inputInstance.Usage.Mem
	instance.LogRateLimit = inputInstance.LogRateLimit
	instance.LogRate = inputInstance.Usage.LogRate
	instance.State = constant.ProcessInstanceState(inputInstance.State)
	instance.Routable = inputInstance.Routable
	instance.Type = inputInstance.Type
	instance.Uptime, err = time.ParseDuration(fmt.Sprintf("%ds", inputInstance.Uptime))
	if err != nil {
		return err
	}

	return nil
}

// DeleteApplicationProcessInstance deletes/stops a particular application's
// process instance.
func (client *Client) DeleteApplicationProcessInstance(appGUID string, processType string, instanceIndex int) (Warnings, error) {
	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName: internal.DeleteApplicationProcessInstanceRequest,
		URIParams: internal.Params{
			"app_guid": appGUID,
			"type":     processType,
			"index":    strconv.Itoa(instanceIndex),
		},
	})

	return warnings, err
}

// GetProcessInstances lists instance stats for a given process.
func (client *Client) GetProcessInstances(processGUID string) ([]ProcessInstance, Warnings, error) {
	var resources []ProcessInstance

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetProcessStatsRequest,
		URIParams:    internal.Params{"process_guid": processGUID},
		ResponseBody: ProcessInstance{},
		AppendToList: func(item interface{}) error {
			resources = append(resources, item.(ProcessInstance))
			return nil
		},
	})

	return resources, warnings, err
}
