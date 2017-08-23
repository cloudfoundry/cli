package v3action

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

// Process represents a V3 actor process.
type Process ccv3.Process

// ProcessSumary represents a process with instance details.
type ProcessSummary struct {
	Process

	InstanceDetails []Instance
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

func (p ProcessSummary) TotalInstanceCount() int {
	return len(p.InstanceDetails)
}

func (p ProcessSummary) HealthyInstanceCount() int {
	count := 0
	for _, instance := range p.InstanceDetails {
		if instance.State == "RUNNING" {
			count++
		}
	}
	return count
}

type ProcessSummaries []ProcessSummary

func (ps ProcessSummaries) Sort() {
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

func (ps ProcessSummaries) String() string {
	ps.Sort()

	var summaries []string
	for _, p := range ps {
		summaries = append(summaries, fmt.Sprintf("%s:%d/%d", p.Type, p.HealthyInstanceCount(), p.TotalInstanceCount()))
	}

	return strings.Join(summaries, ", ")
}

func (actor Actor) ScaleProcessByApplication(appGUID string, process Process) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.CreateApplicationProcessScale(appGUID, ccv3.Process(process))
	allWarnings := Warnings(warnings)
	if err != nil {
		if _, ok := err.(ccerror.ProcessNotFoundError); ok {
			return allWarnings, ProcessNotFoundError{ProcessType: process.Type}
		}
		return allWarnings, err
	}

	return allWarnings, nil
}
