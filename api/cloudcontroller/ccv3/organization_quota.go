package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/types"
)

// OrgQuota represents a Cloud Controller organization quota.
type OrgQuota struct {
	// GUID is the unique ID of the organization quota.
	GUID string `json:"guid,omitempty"`
	// Name is the name of the organization quota
	Name string `json:"name"`
	// Apps contain the various limits that are associated with applications
	Apps AppLimit `json:"apps"`
	// Services contain the various limits that are associated with services
	Services ServiceLimit `json:"services"`
	// Routes contain the various limits that are associated with routes
	Routes RouteLimit `json:"routes"`
}

type AppLimit struct {
	TotalMemory       types.NullInt `json:"total_memory_in_mb"`
	InstanceMemory    types.NullInt `json:"per_process_memory_in_mb"`
	TotalAppInstances types.NullInt `json:"total_instances"`
}

type ServiceLimit struct {
	TotalServiceInstances types.NullInt `json:"total_service_instances"`
	PaidServicePlans      bool          `json:"paid_services_allowed"`
}

type RouteLimit struct {
	TotalRoutes     types.NullInt `json:"total_routes"`
	TotalRoutePorts types.NullInt `json:"total_reserved_ports"`
}

func (client *Client) GetOrganizationQuotas() ([]OrgQuota, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetOrganizationQuotasRequest,
	})
	if err != nil {
		return []OrgQuota{}, nil, err
	}

	var orgQuotasList []OrgQuota
	warnings, err := client.paginate(request, OrgQuota{}, func(item interface{}) error {
		if orgQuota, ok := item.(OrgQuota); ok {
			orgQuotasList = append(orgQuotasList, orgQuota)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   OrgQuota{},
				Unexpected: item,
			}
		}
		return nil
	})

	return orgQuotasList, warnings, err
}
