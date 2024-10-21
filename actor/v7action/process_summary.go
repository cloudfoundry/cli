package v7action

import (
	"fmt"
	"sort"
	"strings"

	"code.cloudfoundry.org/cli/v7/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/v7/resources"
	log "github.com/sirupsen/logrus"
)

// ProcessSummary represents a process with instance details.
type ProcessSummary struct {
	resources.Process

	Sidecars []resources.Sidecar

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

func (actor Actor) getProcessSummariesForApp(appGUID string, withObfuscatedValues bool) (ProcessSummaries, Warnings, error) {
	log.WithFields(log.Fields{
		"appGUID":              appGUID,
		"withObfuscatedValues": withObfuscatedValues,
	}).Info("retrieving process information")

	ccv3Processes, warnings, err := actor.CloudControllerClient.GetApplicationProcesses(appGUID)
	allWarnings := Warnings(warnings)
	if err != nil {
		return nil, allWarnings, err
	}

	var processSummaries ProcessSummaries
	for _, ccv3Process := range ccv3Processes {
		process := resources.Process(ccv3Process)
		if withObfuscatedValues {
			fullProcess, warnings, err := actor.GetProcess(ccv3Process.GUID)
			allWarnings = append(allWarnings, warnings...)
			if err != nil {
				return nil, allWarnings, err
			}
			process = fullProcess
		}

		processSummary, warnings, err := actor.getProcessSummary(process)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return nil, allWarnings, err
		}

		processSummaries = append(processSummaries, processSummary)
	}
	processSummaries.Sort()

	return processSummaries, allWarnings, nil
}

func (actor Actor) getProcessSummary(process resources.Process) (ProcessSummary, Warnings, error) {
	sidecars, warnings, err := actor.CloudControllerClient.GetProcessSidecars(process.GUID)
	allWarnings := Warnings(warnings)
	if err != nil {
		return ProcessSummary{}, allWarnings, err
	}

	instances, warnings, err := actor.CloudControllerClient.GetProcessInstances(process.GUID)
	allWarnings = append(allWarnings, Warnings(warnings)...)
	if err != nil {
		return ProcessSummary{}, allWarnings, err
	}

	processSummary := ProcessSummary{
		Process: process,
	}
	for _, sidecar := range sidecars {
		processSummary.Sidecars = append(processSummary.Sidecars, resources.Sidecar(sidecar))
	}
	for _, instance := range instances {
		processSummary.InstanceDetails = append(processSummary.InstanceDetails, ProcessInstance(instance))
	}

	return processSummary, allWarnings, nil
}
