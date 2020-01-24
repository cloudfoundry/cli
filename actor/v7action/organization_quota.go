package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/types"
)

type OrganizationQuota struct {
	// GUID is the unique ID of the organization quota.
	GUID string `json:"guid,omitempty"`
	// Name is the name of the organization quota
	Name string `json:"name"`

	//the various limits that are associated with applications
	TotalMemory       types.NullInt `json:"total_memory_in_mb"`
	InstanceMemory    types.NullInt `json:"per_process_memory_in_mb"`
	TotalAppInstances types.NullInt `json:"total_instances"`

	//the various limits that are associated with services
	TotalServiceInstances types.NullInt `json:"total_service_instances"`
	PaidServicePlans      bool          `json:"paid_services_allowed"`

	//the various limits that are associated with routes
	TotalRoutes        types.NullInt `json:"total_routes"`
	TotalReservedPorts types.NullInt `json:"total_reserved_ports"`
}

// CreateOrganization creates a new organization with the given name
func (actor Actor) CreateOrganizationQuota(orgQuota OrganizationQuota) (Warnings, error) {
	// Flag that default to 0: total memory, total routes, total reserved ports, total service instances
	setZeroDefaults(&orgQuota)

	if orgQuota.TotalMemory.Value == -1 {
		orgQuota.TotalMemory.IsSet = false
	}

	if orgQuota.InstanceMemory.Value == -1 {
		orgQuota.InstanceMemory.IsSet = false
	}

	if orgQuota.TotalServiceInstances.Value == -1 {
		orgQuota.TotalServiceInstances.IsSet = false
	}

	if orgQuota.TotalAppInstances.Value == -1 {
		orgQuota.TotalAppInstances.IsSet = false
	}

	if orgQuota.TotalRoutes.Value == -1 {
		orgQuota.TotalRoutes.IsSet = false
	}

	if orgQuota.TotalReservedPorts.Value == -1 {
		orgQuota.TotalReservedPorts.IsSet = false
	}

	ccv3Quota := convertToCCV3Quota(orgQuota)
	_, apiWarnings, err := actor.CloudControllerClient.CreateOrganizationQuota(ccv3Quota)

	return Warnings(apiWarnings), err
}

func (actor Actor) GetOrganizationQuotas() ([]OrganizationQuota, Warnings, error) {
	ccv3OrgQuotas, warnings, err := actor.CloudControllerClient.GetOrganizationQuotas()
	if err != nil {
		return []OrganizationQuota{}, Warnings(warnings), err
	}

	var orgQuotas []OrganizationQuota
	for _, quota := range ccv3OrgQuotas {
		orgQuotas = append(orgQuotas, convertToOrganizationQuota(quota))
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
	orgQuota := convertToOrganizationQuota(ccv3OrgQuotas[0])

	return orgQuota, Warnings(warnings), nil
}

func convertToOrganizationQuota(ccv3OrgQuota ccv3.OrgQuota) OrganizationQuota {
	orgQuota := OrganizationQuota{
		GUID:                  ccv3OrgQuota.GUID,
		Name:                  ccv3OrgQuota.Name,
		TotalMemory:           ccv3OrgQuota.Apps.TotalMemory,
		InstanceMemory:        ccv3OrgQuota.Apps.InstanceMemory,
		TotalAppInstances:     ccv3OrgQuota.Apps.TotalAppInstances,
		TotalServiceInstances: ccv3OrgQuota.Services.TotalServiceInstances,
		PaidServicePlans:      ccv3OrgQuota.Services.PaidServicePlans,
		TotalRoutes:           ccv3OrgQuota.Routes.TotalRoutes,
		TotalReservedPorts:    ccv3OrgQuota.Routes.TotalReservedPorts,
	}
	return orgQuota
}

func convertToCCV3Quota(orgQuota OrganizationQuota) ccv3.OrgQuota {
	AppLimit := ccv3.AppLimit{
		TotalMemory:       orgQuota.TotalMemory,
		InstanceMemory:    orgQuota.InstanceMemory,
		TotalAppInstances: orgQuota.TotalAppInstances,
	}
	ServiceLimit := ccv3.ServiceLimit{
		TotalServiceInstances: orgQuota.TotalServiceInstances,
		PaidServicePlans:      orgQuota.PaidServicePlans,
	}
	RouteLimit := ccv3.RouteLimit{
		TotalRoutes:        orgQuota.TotalRoutes,
		TotalReservedPorts: orgQuota.TotalReservedPorts,
	}
	return ccv3.OrgQuota{
		GUID:     orgQuota.GUID,
		Name:     orgQuota.Name,
		Apps:     AppLimit,
		Services: ServiceLimit,
		Routes:   RouteLimit,
	}
}

func setZeroDefaults(orgQuota *OrganizationQuota) {
	orgQuota.TotalMemory.IsSet = true
	orgQuota.TotalRoutes.IsSet = true
	orgQuota.TotalReservedPorts.IsSet = true
	orgQuota.TotalServiceInstances.IsSet = true
}

//func convertUnlimitedToNil(orgQuota OrganizationQuota) {
//	flags := []*types.NullInt{
//		&orgQuota.TotalMemory,
//		&orgQuota.InstanceMemory,
//		&orgQuota.TotalServiceInstances,
//		&orgQuota.TotalAppInstances,
//		&orgQuota.TotalRoutes,
//		&orgQuota.TotalReservedPorts,
//	}
//
//	for i := 0; i < len(flags); i++ {
//		if flags[i].Value == -1 {
//			flags[i].IsSet = false
//		}
//	}
//}
