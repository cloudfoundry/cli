package v3action

import (
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

// Process represents a V3 actor process.
type Process struct {
	Type       string
	Instances  []Instance
	MemoryInMB int
}

// Instance represents a V3 actor instance.
type Instance ccv3.Instance

// StartTime returns the time that the instance started.
func (instance *Instance) StartTime() time.Time {
	uptimeDuration := time.Duration(instance.Uptime) * time.Second

	return time.Now().Add(-uptimeDuration)
}

func (p Process) TotalInstanceCount() int {
	return len(p.Instances)
}

func (p Process) HealthyInstanceCount() int {
	count := 0
	for _, instance := range p.Instances {
		if instance.State == "RUNNING" {
			count++
		}
	}
	return count
}
