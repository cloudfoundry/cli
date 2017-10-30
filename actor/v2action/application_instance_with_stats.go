package v2action

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

type ApplicationInstanceWithStats struct {
	// CPU is the instance's CPU utilization percentage.
	CPU float64

	// Details are arbitrary information about the instance.
	Details string

	// Disk is the instance's disk usage in bytes.
	Disk int

	// DiskQuota is the instance's allowed disk usage in bytes.
	DiskQuota int

	// ID is the instance ID.
	ID int

	// IsolationSegment that the app instance is currently running on.
	IsolationSegment string

	// Memory is the instance's memory usage in bytes.
	Memory int

	// MemoryQuota is the instance's allowed memory usage in bytes.
	MemoryQuota int

	// Since is the Unix time stamp that represents the time the instance was
	// created.
	Since float64

	// State is the instance's state.
	State ApplicationInstanceState
}

// newApplicationInstanceWithStats returns a pointer to a new
// ApplicationInstance.
func newApplicationInstanceWithStats(id int) ApplicationInstanceWithStats {
	return ApplicationInstanceWithStats{ID: id}
}

func (instance ApplicationInstanceWithStats) TimeSinceCreation() time.Time {
	return time.Unix(int64(instance.Since), 0)
}

func (instance *ApplicationInstanceWithStats) setInstance(ccAppInstance ApplicationInstance) {
	instance.Details = ccAppInstance.Details
	instance.Since = ccAppInstance.Since
	instance.State = ApplicationInstanceState(ccAppInstance.State)
}

func (instance *ApplicationInstanceWithStats) setStats(ccAppStats ccv2.ApplicationInstanceStatus) {
	instance.CPU = ccAppStats.CPU
	instance.Disk = ccAppStats.Disk
	instance.DiskQuota = ccAppStats.DiskQuota
	instance.Memory = ccAppStats.Memory
	instance.MemoryQuota = ccAppStats.MemoryQuota
	instance.IsolationSegment = ccAppStats.IsolationSegment
}

func (instance *ApplicationInstanceWithStats) incomplete() {
	instance.Details = strings.TrimSpace(fmt.Sprintf("%s (%s)", instance.Details, "Unable to retrieve information"))
}

func (actor Actor) GetApplicationInstancesWithStatsByApplication(guid string) ([]ApplicationInstanceWithStats, Warnings, error) {
	var allWarnings Warnings

	appInstanceStats, apiWarnings, err := actor.CloudControllerClient.GetApplicationInstanceStatusesByApplication(guid)
	allWarnings = append(allWarnings, apiWarnings...)

	switch err.(type) {
	case ccerror.ResourceNotFoundError, ccerror.ApplicationStoppedStatsError:
		return nil, allWarnings, actionerror.ApplicationInstancesNotFoundError{ApplicationGUID: guid}
	case nil:
		// continue
	default:
		return nil, allWarnings, err
	}

	appInstances, warnings, err := actor.GetApplicationInstancesByApplication(guid)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	returnedInstances := combineStatsAndInstances(appInstanceStats, appInstances)

	sort.Slice(returnedInstances, func(i int, j int) bool { return returnedInstances[i].ID < returnedInstances[j].ID })

	return returnedInstances, allWarnings, err
}

func combineStatsAndInstances(appInstanceStats map[int]ccv2.ApplicationInstanceStatus, appInstances map[int]ApplicationInstance) []ApplicationInstanceWithStats {
	returnedInstances := []ApplicationInstanceWithStats{}
	seenStatuses := make(map[int]bool, len(appInstanceStats))

	for id, appInstanceStat := range appInstanceStats {
		seenStatuses[id] = true

		returnedInstance := newApplicationInstanceWithStats(id)
		returnedInstance.setStats(appInstanceStat)

		if appInstance, found := appInstances[id]; found {
			returnedInstance.setInstance(appInstance)
		} else {
			returnedInstance.incomplete()
		}

		returnedInstances = append(returnedInstances, returnedInstance)
	}

	// add instances that are missing stats
	for index, appInstance := range appInstances {
		if _, found := seenStatuses[index]; !found {
			returnedInstance := newApplicationInstanceWithStats(index)
			returnedInstance.setInstance(appInstance)
			returnedInstance.incomplete()

			returnedInstances = append(returnedInstances, returnedInstance)
		}
	}

	return returnedInstances
}
