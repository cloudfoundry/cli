package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/api/cloudcontroller/jsonry"
)

// ServicePlan represents a Cloud Controller V3 Service Plan.
type ServicePlan struct {
	// GUID is a unique service plan identifier.
	GUID string
	// Name is the name of the service plan.
	Name string
	// VisibilityType can be "public", "admin", "organization" or "space"
	VisibilityType VisibilityType `json:"visibility_type"`
	// ServicePlanGUID is the GUID of the service offering
	ServiceOfferingGUID string `jsonry:"relationships.service_offering.data.guid"`
	// SpaceGUID is the space that a plan from a space-scoped broker relates to
	SpaceGUID string `jsonry:"relationships.space.data.guid"`

	Metadata *Metadata
}

func (sp *ServicePlan) UnmarshalJSON(data []byte) error {
	return jsonry.Unmarshal(data, sp)
}

// GetServicePlans lists service plan with optional filters.
func (client *Client) GetServicePlans(query ...Query) ([]ServicePlan, Warnings, error) {
	plans, _, warnings, err := client.getServicePlans(query...)
	return plans, warnings, err
}

func (client *Client) getServicePlans(query ...Query) ([]ServicePlan, IncludedResources, Warnings, error) {
	var plans []ServicePlan

	included, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetServicePlansRequest,
		Query:        query,
		ResponseBody: ServicePlan{},
		AppendToList: func(item interface{}) error {
			plans = append(plans, item.(ServicePlan))
			return nil
		},
	})

	return plans, included, warnings, err
}

type ServicePlanWithSpaceAndOrganization struct {
	// GUID is a unique service plan identifier.
	GUID string
	// Name is the name of the service plan.
	Name string
	// VisibilityType can be "public", "admin", "organization" or "space"
	VisibilityType VisibilityType
	// ServicePlanGUID is the GUID of the service offering
	ServiceOfferingGUID string

	SpaceGUID string

	SpaceName string

	OrganizationName string
}

type planSpaceDetails struct{ spaceName, orgName string }

func (client *Client) GetServicePlansWithSpaceAndOrganization(query ...Query) ([]ServicePlanWithSpaceAndOrganization, Warnings, error) {
	query = append(query, Query{
		Key:    Include,
		Values: []string{"space.organization"},
	})

	plans, included, warnings, err := client.getServicePlans(query...)

	spaceDetailsFromGUID := computeSpaceDetailsTable(included)

	var enrichedPlans []ServicePlanWithSpaceAndOrganization
	for _, plan := range plans {
		sd := spaceDetailsFromGUID[plan.SpaceGUID]

		enrichedPlans = append(enrichedPlans, ServicePlanWithSpaceAndOrganization{
			GUID:                plan.GUID,
			Name:                plan.Name,
			VisibilityType:      plan.VisibilityType,
			ServiceOfferingGUID: plan.ServiceOfferingGUID,
			SpaceGUID:           plan.SpaceGUID,
			SpaceName:           sd.spaceName,
			OrganizationName:    sd.orgName,
		})
	}

	return enrichedPlans, warnings, err
}

func computeSpaceDetailsTable(included IncludedResources) map[string]planSpaceDetails {
	orgNameFromGUID := make(map[string]string)
	for _, org := range included.Organizations {
		orgNameFromGUID[org.GUID] = org.Name
	}

	spaceDetailsFromGUID := make(map[string]planSpaceDetails)
	for _, space := range included.Spaces {
		details := planSpaceDetails{spaceName: space.Name}

		if orgRelationship, ok := space.Relationships[constant.RelationshipTypeOrganization]; ok {
			details.orgName = orgNameFromGUID[orgRelationship.GUID]
		}

		spaceDetailsFromGUID[space.GUID] = details
	}

	return spaceDetailsFromGUID
}
