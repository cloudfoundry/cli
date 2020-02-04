package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/types"
)

type OrganizationQuota ccv3.OrganizationQuota

type QuotaLimits struct {
	TotalMemoryInMB       *types.NullInt
	PerProcessMemoryInMB  *types.NullInt
	TotalInstances        *types.NullInt
	PaidServicesAllowed   *bool
	TotalServiceInstances *types.NullInt
	TotalRoutes           *types.NullInt
	TotalReservedPorts    *types.NullInt
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

func (actor Actor) GetOrganizationQuotas() ([]OrganizationQuota, Warnings, error) {
	ccv3OrgQuotas, warnings, err := actor.CloudControllerClient.GetOrganizationQuotas()
	if err != nil {
		return []OrganizationQuota{}, Warnings(warnings), err
	}

	var orgQuotas []OrganizationQuota
	for _, quota := range ccv3OrgQuotas {
		orgQuotas = append(orgQuotas, OrganizationQuota(quota))
	}

	return orgQuotas, Warnings(warnings), nil
}

func (actor Actor) GetOrganizationQuotaByName(orgQuotaName string) (OrganizationQuota, Warnings, error) {
	ccv3OrgQuotas, warnings, err := actor.CloudControllerClient.GetOrganizationQuotas(
		ccv3.Query{
			Key:    ccv3.NameFilter,
			Values: []string{orgQuotaName},
		},
	)
	if err != nil {
		return OrganizationQuota{}, Warnings(warnings), err

	}

	if len(ccv3OrgQuotas) == 0 {
		return OrganizationQuota{}, Warnings(warnings), actionerror.OrganizationQuotaNotFoundForNameError{Name: orgQuotaName}
	}
	orgQuota := OrganizationQuota(ccv3OrgQuotas[0])

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

func createQuotaStruct(name string, limits QuotaLimits) ccv3.OrganizationQuota {
	AppLimit := ccv3.AppLimit{
		TotalMemory:       limits.TotalMemoryInMB,
		InstanceMemory:    limits.PerProcessMemoryInMB,
		TotalAppInstances: limits.TotalInstances,
	}
	ServiceLimit := ccv3.ServiceLimit{
		TotalServiceInstances: limits.TotalServiceInstances,
		PaidServicePlans:      limits.PaidServicesAllowed,
	}
	RouteLimit := ccv3.RouteLimit{
		TotalRoutes:        limits.TotalRoutes,
		TotalReservedPorts: limits.TotalReservedPorts,
	}

	quota := ccv3.OrganizationQuota{
		Name:     name,
		Apps:     AppLimit,
		Services: ServiceLimit,
		Routes:   RouteLimit,
	}

	return quota
}

func setZeroDefaultsForQuotaCreation(apps *ccv3.AppLimit, routes *ccv3.RouteLimit, services *ccv3.ServiceLimit) {
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

func convertUnlimitedToNil(apps *ccv3.AppLimit, routes *ccv3.RouteLimit, services *ccv3.ServiceLimit) {
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
