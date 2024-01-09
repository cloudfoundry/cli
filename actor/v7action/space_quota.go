package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
)

func (actor Actor) ApplySpaceQuotaByName(quotaName, spaceGUID, orgGUID string) (Warnings, error) {
	var allWarnings Warnings

	spaceQuota, warnings, err := actor.GetSpaceQuotaByName(quotaName, orgGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	_, ccWarnings, err := actor.CloudControllerClient.ApplySpaceQuota(spaceQuota.GUID, spaceGUID)
	allWarnings = append(allWarnings, ccWarnings...)

	return allWarnings, err
}

func (actor Actor) CreateSpaceQuota(spaceQuotaName string, orgGuid string, limits QuotaLimits) (Warnings, error) {
	allWarnings := Warnings{}

	spaceQuota := resources.SpaceQuota{
		Quota: resources.Quota{
			Name: spaceQuotaName,
			Apps: resources.AppLimit{
				TotalMemory:       limits.TotalMemoryInMB,
				InstanceMemory:    limits.PerProcessMemoryInMB,
				TotalAppInstances: limits.TotalInstances,
				TotalLogVolume:    limits.TotalLogVolume,
			},
			Services: resources.ServiceLimit{
				TotalServiceInstances: limits.TotalServiceInstances,
				PaidServicePlans:      limits.PaidServicesAllowed,
			},
			Routes: resources.RouteLimit{
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

func (actor Actor) DeleteSpaceQuotaByName(quotaName string, orgGUID string) (Warnings, error) {
	var allWarnings Warnings

	quota, warnings, err := actor.GetSpaceQuotaByName(quotaName, orgGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	jobURL, ccv3Warnings, err := actor.CloudControllerClient.DeleteSpaceQuota(quota.GUID)
	allWarnings = append(allWarnings, Warnings(ccv3Warnings)...)
	if err != nil {
		return allWarnings, err
	}

	ccv3Warnings, err = actor.CloudControllerClient.PollJob(jobURL)
	allWarnings = append(allWarnings, Warnings(ccv3Warnings)...)
	if err != nil {
		return allWarnings, err
	}

	return allWarnings, nil
}

func (actor Actor) GetSpaceQuotaByName(spaceQuotaName string, orgGUID string) (resources.SpaceQuota, Warnings, error) {
	ccv3Quotas, warnings, err := actor.CloudControllerClient.GetSpaceQuotas(
		ccv3.Query{Key: ccv3.OrganizationGUIDFilter, Values: []string{orgGUID}},
		ccv3.Query{Key: ccv3.NameFilter, Values: []string{spaceQuotaName}},
		ccv3.Query{Key: ccv3.PerPage, Values: []string{"1"}},
		ccv3.Query{Key: ccv3.Page, Values: []string{"1"}},
	)

	if err != nil {
		return resources.SpaceQuota{}, Warnings(warnings), err
	}

	if len(ccv3Quotas) == 0 {
		return resources.SpaceQuota{}, Warnings(warnings), actionerror.SpaceQuotaNotFoundForNameError{Name: spaceQuotaName}
	}

	return ccv3Quotas[0], Warnings(warnings), nil
}

func (actor Actor) GetSpaceQuotasByOrgGUID(orgGUID string) ([]resources.SpaceQuota, Warnings, error) {
	spaceQuotas, warnings, err := actor.CloudControllerClient.GetSpaceQuotas(
		ccv3.Query{
			Key:    ccv3.OrganizationGUIDFilter,
			Values: []string{orgGUID},
		},
	)

	if err != nil {
		return []resources.SpaceQuota{}, Warnings(warnings), err
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

	newSpaceQuota := resources.SpaceQuota{
		Quota: resources.Quota{
			GUID: oldSpaceQuota.GUID,
			Name: newName,
			Apps: resources.AppLimit{
				TotalMemory:       limits.TotalMemoryInMB,
				InstanceMemory:    limits.PerProcessMemoryInMB,
				TotalAppInstances: limits.TotalInstances,
				TotalLogVolume:    limits.TotalLogVolume,
			},
			Services: resources.ServiceLimit{
				TotalServiceInstances: limits.TotalServiceInstances,
				PaidServicePlans:      limits.PaidServicesAllowed,
			},
			Routes: resources.RouteLimit{
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

func (actor Actor) UnsetSpaceQuota(spaceQuotaName, spaceName, orgGUID string) (Warnings, error) {
	var allWarnings Warnings
	space, warnings, err := actor.GetSpaceByNameAndOrganization(spaceName, orgGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	spaceQuota, warnings, err := actor.GetSpaceQuotaByName(spaceQuotaName, orgGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	ccWarnings, err := actor.CloudControllerClient.UnsetSpaceQuota(spaceQuota.GUID, space.GUID)
	allWarnings = append(allWarnings, ccWarnings...)
	if err != nil {
		return allWarnings, err
	}

	return allWarnings, nil
}
