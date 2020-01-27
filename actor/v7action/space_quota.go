package v7action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/types"
)

type SpaceQuota struct {
	// GUID is the unique ID of the space  quota.
	GUID string
	// Name is the name of the space quota
	Name string

	//the various limits that are associated with applications
	TotalMemoryInMB      types.NullInt
	PerProcessMemoryInMB types.NullInt
	TotalInstances       types.NullInt
	PerAppTasks          types.NullInt

	//the various limits that are associated with services
	PaidServicesAllowed   bool
	TotalServiceInstances types.NullInt
	TotalServiceKeys      types.NullInt

	//the various limits that are associated with routes
	TotalRoutes        types.NullInt
	TotalReservedPorts types.NullInt
}

type SpaceQuotaLimits struct {
	TotalMemoryInMB       types.NullInt
	PerProcessMemoryInMB  types.NullInt
	TotalInstances        types.NullInt
	PerAppTasks           types.NullInt
	PaidServicesAllowed   bool
	TotalServiceInstances types.NullInt
	TotalServiceKeys      types.NullInt
	TotalRoutes           types.NullInt
	TotalReservedPorts    types.NullInt
}

func (actor Actor) CreateSpaceQuota(spaceQuotaName string, orgGuid string, limits SpaceQuotaLimits) (Warnings, error) {
	allWarnings := Warnings{}

	spaceQuota := ccv3.SpaceQuota{
		Name: spaceQuotaName,
		Apps: ccv3.AppLimit{
			TotalMemory:       limits.TotalMemoryInMB,
			InstanceMemory:    limits.PerProcessMemoryInMB,
			TotalAppInstances: limits.TotalInstances,
		},
		Services: ccv3.ServiceLimit{
			TotalServiceInstances: limits.TotalServiceInstances,
			PaidServicePlans:      limits.PaidServicesAllowed,
		},
		Routes: ccv3.RouteLimit{
			TotalRoutes:     limits.TotalRoutes,
			TotalRoutePorts: limits.TotalReservedPorts,
		},
		OrgGUID:    orgGuid,
		SpaceGUIDs: nil,
	}

	_, warnings, err := actor.CloudControllerClient.CreateSpaceQuota(spaceQuota)
	allWarnings = append(allWarnings, warnings...)

	return allWarnings, err
}
