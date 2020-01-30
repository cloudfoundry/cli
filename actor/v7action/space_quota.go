package v7action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

func (actor Actor) CreateSpaceQuota(spaceQuotaName string, orgGuid string, limits QuotaLimits) (Warnings, error) {
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
			TotalRoutes:        limits.TotalRoutes,
			TotalReservedPorts: limits.TotalReservedPorts,
		},
		OrgGUID:    orgGuid,
		SpaceGUIDs: nil,
	}

	setZeroDefaultsForQuotaCreation(&spaceQuota.Apps, &spaceQuota.Routes, &spaceQuota.Services)
	convertUnlimitedToNil(&spaceQuota.Apps, &spaceQuota.Routes, &spaceQuota.Services)

	_, warnings, err := actor.CloudControllerClient.CreateSpaceQuota(spaceQuota)
	allWarnings = append(allWarnings, warnings...)

	return allWarnings, err
}
