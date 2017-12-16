package v3action

import (
	"fmt"
	"sort"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
)

// ProcessSummary represents a process with instance details.
type ProcessSummary struct {
	Process

	InstanceDetails []ProcessInstance
}

type ProcessSummaries []ProcessSummary

func (p ProcessSummary) TotalInstanceCount() int {
	return len(p.InstanceDetails)
}

func (p ProcessSummary) HealthyInstanceCount() int {
	count := 0
	for _, instance := range p.InstanceDetails {
		if instance.State == constant.ProcessInstanceRunning {
			count++
		}
	}
	return count
}

func (ps ProcessSummaries) Sort() {
	sort.Slice(ps, func(i int, j int) bool {
		var iScore int
		var jScore int

		switch ps[i].Type {
		case constant.ProcessTypeWeb:
			iScore = 0
		default:
			iScore = 1
		}

		switch ps[j].Type {
		case constant.ProcessTypeWeb:
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

func (actor Actor) getProcessSummariesForApp(appGUID string) (ProcessSummaries, Warnings, error) {
	var allWarnings Warnings

	ccv3Processes, warnings, err := actor.CloudControllerClient.GetApplicationProcesses(appGUID)
	allWarnings = Warnings(warnings)
	if err != nil {
		return nil, allWarnings, err
	}

	var processSummaries ProcessSummaries
	for _, ccv3Process := range ccv3Processes {
		processGUID := ccv3Process.GUID
		instances, warnings, err := actor.CloudControllerClient.GetProcessInstances(processGUID)
		allWarnings = append(allWarnings, Warnings(warnings)...)
		if err != nil {
			return nil, allWarnings, err
		}

		processSummary := ProcessSummary{
			Process: Process(ccv3Process),
		}
		for _, instance := range instances {
			processSummary.InstanceDetails = append(processSummary.InstanceDetails, ProcessInstance(instance))
		}

		processSummaries = append(processSummaries, processSummary)
	}

	return processSummaries, allWarnings, nil
}
