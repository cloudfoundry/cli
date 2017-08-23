package v3action

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/types"
)

// Process represents a V3 actor process.
type Process struct {
	Type                  string
	Instances             []Instance
	DesiredInstancesCount types.NullInt
	MemoryInMB            types.NullUint64
	DiskInMB              types.NullUint64
}

// Instance represents a V3 actor instance.
type Instance ccv3.Instance

// ProcessNotFoundError is returned when the proccess type cannot be found
type ProcessNotFoundError struct {
	ProcessType string
}

func (e ProcessNotFoundError) Error() string {
	return fmt.Sprintf("Process %s not found", e.ProcessType)
}

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

type Processes []Process

func (ps Processes) Sort() {
	sort.Slice(ps, func(i int, j int) bool {
		var iScore int
		var jScore int

		switch ps[i].Type {
		case "web":
			iScore = 0
		default:
			iScore = 1
		}

		switch ps[j].Type {
		case "web":
			jScore = 0
		default:
			jScore = 1
		}

		if iScore == 1 && jScore == 1 {
			return ps[i].Type < ps[j].Type
		}
		return iScore < jScore
	})

}

func (ps Processes) Summary() string {
	ps.Sort()

	var summaries []string
	for _, p := range ps {
		summaries = append(summaries, fmt.Sprintf("%s:%d/%d", p.Type, p.HealthyInstanceCount(), p.TotalInstanceCount()))
	}

	return strings.Join(summaries, ", ")
}

func (actor Actor) ScaleProcessByApplication(appGUID string, process Process) (Warnings, error) {
	ccv3Process := ccv3.Process{
		Type:       process.Type,
		Instances:  process.DesiredInstancesCount,
		MemoryInMB: process.MemoryInMB,
		DiskInMB:   process.DiskInMB,
	}
	warnings, err := actor.CloudControllerClient.CreateApplicationProcessScale(appGUID, ccv3Process)
	allWarnings := Warnings(warnings)
	if err != nil {
		if _, ok := err.(ccerror.ProcessNotFoundError); ok {
			return allWarnings, ProcessNotFoundError{ProcessType: process.Type}
		}
		return allWarnings, err
	}

	return allWarnings, nil
}
