package resources

import (
	"encoding/json"
	"fmt"

	"code.cloudfoundry.org/cli/v9/api/cloudcontroller"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/v9/types"
)

type Process struct {
	GUID string
	Type string
	// Command is the process start command. Note: This value will be obfuscated when obtained from listing.
	Command                               types.FilteredString
	HealthCheckType                       constant.HealthCheckType
	HealthCheckEndpoint                   string
	HealthCheckInvocationTimeout          int64
	HealthCheckTimeout                    int64
	ReadinessHealthCheckType              constant.HealthCheckType
	ReadinessHealthCheckEndpoint          string
	ReadinessHealthCheckInvocationTimeout int64
	ReadinessHealthCheckInterval          int64
	Instances                             types.NullInt
	MemoryInMB                            types.NullUint64
	DiskInMB                              types.NullUint64
	LogRateLimitInBPS                     types.NullInt
	AppGUID                               string
}

func (p Process) MarshalJSON() ([]byte, error) {
	var ccProcess marshalProcess

	marshalCommand(p, &ccProcess)
	marshalInstances(p, &ccProcess)
	marshalMemory(p, &ccProcess)
	marshalDisk(p, &ccProcess)
	marshalLogRateLimit(p, &ccProcess)
	marshalHealthCheck(p, &ccProcess)
	marshalReadinessHealthCheck(p, &ccProcess)

	return json.Marshal(ccProcess)
}

func (p *Process) UnmarshalJSON(data []byte) error {
	var ccProcess struct {
		Command           types.FilteredString `json:"command"`
		DiskInMB          types.NullUint64     `json:"disk_in_mb"`
		GUID              string               `json:"guid"`
		Instances         types.NullInt        `json:"instances"`
		MemoryInMB        types.NullUint64     `json:"memory_in_mb"`
		LogRateLimitInBPS types.NullInt        `json:"log_rate_limit_in_bytes_per_second"`
		Type              string               `json:"type"`
		Relationships     Relationships        `json:"relationships"`

		HealthCheck struct {
			Type constant.HealthCheckType `json:"type"`
			Data struct {
				Endpoint          string `json:"endpoint"`
				InvocationTimeout int64  `json:"invocation_timeout"`
				Timeout           int64  `json:"timeout"`
			} `json:"data"`
		} `json:"health_check"`

		ReadinessHealthCheck struct {
			Type constant.HealthCheckType `json:"type"`
			Data struct {
				Endpoint          string `json:"endpoint"`
				InvocationTimeout int64  `json:"invocation_timeout"`
				Interval          int64  `json:"interval"`
			} `json:"data"`
		} `json:"readiness_health_check"`
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
	p.HealthCheckTimeout = ccProcess.HealthCheck.Data.Timeout
	p.HealthCheckType = ccProcess.HealthCheck.Type
	p.ReadinessHealthCheckEndpoint = ccProcess.ReadinessHealthCheck.Data.Endpoint
	p.ReadinessHealthCheckType = ccProcess.ReadinessHealthCheck.Type
	p.ReadinessHealthCheckInvocationTimeout = ccProcess.ReadinessHealthCheck.Data.InvocationTimeout
	p.ReadinessHealthCheckInterval = ccProcess.ReadinessHealthCheck.Data.Interval
	p.Instances = ccProcess.Instances
	p.MemoryInMB = ccProcess.MemoryInMB
	p.LogRateLimitInBPS = ccProcess.LogRateLimitInBPS
	p.Type = ccProcess.Type
	p.AppGUID = ccProcess.Relationships[constant.RelationshipTypeApplication].GUID

	return nil
}

type healthCheck struct {
	Type constant.HealthCheckType `json:"type,omitempty"`
	Data struct {
		Endpoint          interface{} `json:"endpoint,omitempty"`
		InvocationTimeout int64       `json:"invocation_timeout,omitempty"`
		Timeout           int64       `json:"timeout,omitempty"`
	} `json:"data"`
}

type readinessHealthCheck struct {
	Type constant.HealthCheckType `json:"type,omitempty"`
	Data struct {
		Endpoint          interface{} `json:"endpoint,omitempty"`
		InvocationTimeout int64       `json:"invocation_timeout,omitempty"`
		Interval          int64       `json:"interval,omitempty"`
	} `json:"data"`
}

type marshalProcess struct {
	Command           interface{} `json:"command,omitempty"`
	Instances         json.Number `json:"instances,omitempty"`
	MemoryInMB        json.Number `json:"memory_in_mb,omitempty"`
	DiskInMB          json.Number `json:"disk_in_mb,omitempty"`
	LogRateLimitInBPS json.Number `json:"log_rate_limit_in_bytes_per_second,omitempty"`

	HealthCheck          *healthCheck          `json:"health_check,omitempty"`
	ReadinessHealthCheck *readinessHealthCheck `json:"readiness_health_check,omitempty"`
}

func marshalCommand(p Process, ccProcess *marshalProcess) {
	if p.Command.IsSet {
		ccProcess.Command = &p.Command
	}
}

func marshalDisk(p Process, ccProcess *marshalProcess) {
	if p.DiskInMB.IsSet {
		ccProcess.DiskInMB = json.Number(fmt.Sprint(p.DiskInMB.Value))
	}
}

func marshalHealthCheck(p Process, ccProcess *marshalProcess) {
	if p.HealthCheckType != "" || p.HealthCheckEndpoint != "" || p.HealthCheckInvocationTimeout != 0 || p.HealthCheckTimeout != 0 {
		ccProcess.HealthCheck = new(healthCheck)
		ccProcess.HealthCheck.Type = p.HealthCheckType
		ccProcess.HealthCheck.Data.InvocationTimeout = p.HealthCheckInvocationTimeout
		ccProcess.HealthCheck.Data.Timeout = p.HealthCheckTimeout
		if p.HealthCheckEndpoint != "" {
			ccProcess.HealthCheck.Data.Endpoint = p.HealthCheckEndpoint
		}
	}
}

func marshalReadinessHealthCheck(p Process, ccProcess *marshalProcess) {
	if p.ReadinessHealthCheckType != "" || p.ReadinessHealthCheckEndpoint != "" || p.ReadinessHealthCheckInvocationTimeout != 0 {
		ccProcess.ReadinessHealthCheck = new(readinessHealthCheck)
		ccProcess.ReadinessHealthCheck.Type = p.ReadinessHealthCheckType
		ccProcess.ReadinessHealthCheck.Data.InvocationTimeout = p.ReadinessHealthCheckInvocationTimeout
		ccProcess.ReadinessHealthCheck.Data.Interval = p.ReadinessHealthCheckInterval
		if p.ReadinessHealthCheckEndpoint != "" {
			ccProcess.ReadinessHealthCheck.Data.Endpoint = p.ReadinessHealthCheckEndpoint
		}
	}
}

func marshalInstances(p Process, ccProcess *marshalProcess) {
	if p.Instances.IsSet {
		ccProcess.Instances = json.Number(fmt.Sprint(p.Instances.Value))
	}
}

func marshalMemory(p Process, ccProcess *marshalProcess) {
	if p.MemoryInMB.IsSet {
		ccProcess.MemoryInMB = json.Number(fmt.Sprint(p.MemoryInMB.Value))
	}
}

func marshalLogRateLimit(p Process, ccProcess *marshalProcess) {
	if p.LogRateLimitInBPS.IsSet {
		ccProcess.LogRateLimitInBPS = json.Number(fmt.Sprint(p.LogRateLimitInBPS.Value))
	}
}
