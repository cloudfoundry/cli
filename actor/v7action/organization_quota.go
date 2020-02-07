package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
)

type QuotaLimits struct {
	TotalMemoryInMB       *types.NullInt
	PerProcessMemoryInMB  *types.NullInt
	TotalInstances        *types.NullInt
	PaidServicesAllowed   *bool
	TotalServiceInstances *types.NullInt
	TotalRoutes           *types.NullInt
	TotalReservedPorts    *types.NullInt
}

func (actor Actor) ApplyOrganizationQuotaByName(quotaName string, orgGUID string) (Warnings, error) {
	var allWarnings Warnings

	quota, warnings, err := actor.GetOrganizationQuotaByName(quotaName)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	_, ccWarnings, err := actor.CloudControllerClient.ApplyOrganizationQuota(quota.GUID, orgGUID)
	allWarnings = append(allWarnings, ccWarnings...)

	return allWarnings, err
}

// CreateOrganization creates a new organization with the given name
func (actor Actor) CreateOrganizationQuota(name string, limits QuotaLimits) (Warnings, error) {
	orgQuota := createQuotaStruct(name, limits)
	setZeroDefaultsForQuotaCreation(&orgQuota.Apps, &orgQuota.Routes, &orgQuota.Services)
	convertUnlimitedToNil(&orgQuota.Apps, &orgQuota.Routes, &orgQuota.Services)

	_, apiWarnings, err := actor.CloudControllerClient.CreateOrganizationQuota(orgQuota)

	return Warnings(apiWarnings), err
}

func (actor Actor) DeleteOrganizationQuota(quotaName string) (Warnings, error) {
	var allWarnings Warnings

	quota, warnings, err := actor.GetOrganizationQuotaByName(quotaName)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	jobURL, ccWarnings, err := actor.CloudControllerClient.DeleteOrganizationQuota(quota.GUID)
	allWarnings = append(allWarnings, ccWarnings...)
	if err != nil {
		return allWarnings, err
	}

	ccWarnings, err = actor.CloudControllerClient.PollJob(jobURL)
	allWarnings = append(allWarnings, ccWarnings...)

	return allWarnings, err
}

func (actor Actor) GetOrganizationQuotas() ([]resources.OrganizationQuota, Warnings, error) {
	orgQuotas, warnings, err := actor.CloudControllerClient.GetOrganizationQuotas()
	if err != nil {
		return []resources.OrganizationQuota{}, Warnings(warnings), err
	}

	return orgQuotas, Warnings(warnings), nil
}

func (actor Actor) GetOrganizationQuotaByName(orgQuotaName string) (resources.OrganizationQuota, Warnings, error) {
	ccv3OrgQuotas, warnings, err := actor.CloudControllerClient.GetOrganizationQuotas(
		ccv3.Query{
			Key:    ccv3.NameFilter,
			Values: []string{orgQuotaName},
		},
	)
	if err != nil {
		return resources.OrganizationQuota{}, Warnings(warnings), err

	}

	if len(ccv3OrgQuotas) == 0 {
		return resources.OrganizationQuota{}, Warnings(warnings), actionerror.OrganizationQuotaNotFoundForNameError{Name: orgQuotaName}
	}
	orgQuota := ccv3OrgQuotas[0]

	return orgQuota, Warnings(warnings), nil
}

func (actor Actor) UpdateOrganizationQuota(quotaName string, newName string, limits QuotaLimits) (Warnings, error) {
	var allWarnings Warnings

	quota, warnings, err := actor.GetOrganizationQuotaByName(quotaName)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	if newName == "" {
		newName = quotaName
	}
	newQuota := createQuotaStruct(newName, limits)
	newQuota.GUID = quota.GUID
	convertUnlimitedToNil(&newQuota.Apps, &newQuota.Routes, &newQuota.Services)

	_, ccWarnings, err := actor.CloudControllerClient.UpdateOrganizationQuota(newQuota)
	allWarnings = append(allWarnings, ccWarnings...)

	return allWarnings, err
}

func createQuotaStruct(name string, limits QuotaLimits) resources.OrganizationQuota {
	AppLimit := resources.AppLimit{
		TotalMemory:       limits.TotalMemoryInMB,
		InstanceMemory:    limits.PerProcessMemoryInMB,
		TotalAppInstances: limits.TotalInstances,
	}
	ServiceLimit := resources.ServiceLimit{
		TotalServiceInstances: limits.TotalServiceInstances,
		PaidServicePlans:      limits.PaidServicesAllowed,
	}
	RouteLimit := resources.RouteLimit{
		TotalRoutes:        limits.TotalRoutes,
		TotalReservedPorts: limits.TotalReservedPorts,
	}

	quota := resources.OrganizationQuota{
		Quota: resources.Quota{
			Name:     name,
			Apps:     AppLimit,
			Services: ServiceLimit,
			Routes:   RouteLimit,
		},
	}

	return quota
}

func setZeroDefaultsForQuotaCreation(apps *resources.AppLimit, routes *resources.RouteLimit, services *resources.ServiceLimit) {
	if apps.TotalMemory == nil {
		apps.TotalMemory = &types.NullInt{IsSet: true, Value: 0}
	}

	if routes.TotalRoutes == nil {
		routes.TotalRoutes = &types.NullInt{IsSet: true, Value: 0}
	}

	if routes.TotalReservedPorts == nil {
		routes.TotalReservedPorts = &types.NullInt{IsSet: true, Value: 0}
	}

	if services.TotalServiceInstances == nil {
		services.TotalServiceInstances = &types.NullInt{IsSet: true, Value: 0}
	}
}

func convertUnlimitedToNil(apps *resources.AppLimit, routes *resources.RouteLimit, services *resources.ServiceLimit) {
	flags := []*types.NullInt{
		apps.TotalMemory,
		apps.InstanceMemory,
		apps.TotalAppInstances,
		services.TotalServiceInstances,
		routes.TotalRoutes,
		routes.TotalReservedPorts,
	}

	for i := 0; i < len(flags); i++ {
		if flags[i] != nil {
			if flags[i].Value == -1 {
				flags[i].IsSet = false
				flags[i].Value = 0
			}
		}
	}
}
