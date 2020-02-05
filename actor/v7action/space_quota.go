package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

type SpaceQuota ccv3.SpaceQuota

func (actor Actor) CreateSpaceQuota(spaceQuotaName string, orgGuid string, limits QuotaLimits) (Warnings, error) {
	allWarnings := Warnings{}

	spaceQuota := ccv3.SpaceQuota{
		Quota: ccv3.Quota{
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

func (actor Actor) GetSpaceQuotaByName(spaceQuotaName string, orgGUID string) (SpaceQuota, Warnings, error) {
	ccv3Quotas, warnings, err := actor.CloudControllerClient.GetSpaceQuotas(
		ccv3.Query{
			Key:    ccv3.OrganizationGUIDFilter,
			Values: []string{orgGUID},
		},
		ccv3.Query{
			Key:    ccv3.NameFilter,
			Values: []string{spaceQuotaName},
		},
	)

	if err != nil {
		return SpaceQuota{}, Warnings(warnings), err
	}

	if len(ccv3Quotas) == 0 {
		return SpaceQuota{}, Warnings(warnings), actionerror.SpaceQuotaNotFoundByNameError{Name: spaceQuotaName}
	}

	return SpaceQuota(ccv3Quotas[0]), Warnings(warnings), nil
}

func (actor Actor) GetSpaceQuotasByOrgGUID(orgGUID string) ([]SpaceQuota, Warnings, error) {
	ccv3Quotas, warnings, err := actor.CloudControllerClient.GetSpaceQuotas(
		ccv3.Query{
			Key:    ccv3.OrganizationGUIDFilter,
			Values: []string{orgGUID},
		},
	)

	if err != nil {
		return []SpaceQuota{}, Warnings(warnings), err
	}

	var spaceQuotas []SpaceQuota
	for _, quota := range ccv3Quotas {
		spaceQuotas = append(spaceQuotas, SpaceQuota(quota))
	}

	return spaceQuotas, Warnings(warnings), nil
}

func (actor Actor) UpdateSpaceQuota(currentName, orgGUID, newName string, limits QuotaLimits) (Warnings, error) {
	var allWarnings Warnings

	oldSpaceQuota, warnings, err := actor.GetSpaceQuotaByName(currentName, orgGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	if newName == "" {
		newName = currentName
	}

	newSpaceQuota := ccv3.SpaceQuota{
		Quota: ccv3.Quota{
			GUID: oldSpaceQuota.GUID,
			Name: newName,
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
		},
	}

	convertUnlimitedToNil(&newSpaceQuota.Apps, &newSpaceQuota.Routes, &newSpaceQuota.Services)

	_, ccWarnings, err := actor.CloudControllerClient.UpdateSpaceQuota(newSpaceQuota)
	allWarnings = append(allWarnings, ccWarnings...)

	return allWarnings, err
}
