package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/types"
)

type OrganizationQuota ccv3.OrganizationQuota

type QuotaLimits struct {
	TotalMemoryInMB       types.NullInt
	PerProcessMemoryInMB  types.NullInt
	TotalInstances        types.NullInt
	PaidServicesAllowed   bool
	TotalServiceInstances types.NullInt
	TotalRoutes           types.NullInt
	TotalReservedPorts    types.NullInt
}

// CreateOrganization creates a new organization with the given name
func (actor Actor) CreateOrganizationQuota(name string, limits QuotaLimits) (Warnings, error) {
	orgQuota := createQuotaStruct(name, limits)

	_, apiWarnings, err := actor.CloudControllerClient.CreateOrganizationQuota(orgQuota)

	return Warnings(apiWarnings), err
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

	setZeroDefaults(&quota)
	convertUnlimitedToNil(&quota)

	return quota
}

func setZeroDefaults(orgQuota *ccv3.OrganizationQuota) {
	orgQuota.Apps.TotalMemory.IsSet = true
	orgQuota.Routes.TotalRoutes.IsSet = true
	orgQuota.Routes.TotalReservedPorts.IsSet = true
	orgQuota.Services.TotalServiceInstances.IsSet = true
}

func convertUnlimitedToNil(orgQuota *ccv3.OrganizationQuota) {
	flags := []*types.NullInt{
		&orgQuota.Apps.TotalMemory,
		&orgQuota.Apps.InstanceMemory,
		&orgQuota.Apps.TotalAppInstances,
		&orgQuota.Services.TotalServiceInstances,
		&orgQuota.Routes.TotalRoutes,
		&orgQuota.Routes.TotalReservedPorts,
	}

	for i := 0; i < len(flags); i++ {
		if flags[i].Value == -1 {
			flags[i].IsSet = false
		}
	}
}
