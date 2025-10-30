package ccv3

import (
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/v8/resources"
	"code.cloudfoundry.org/cli/v8/util/lookuptable"
)

func (client *Client) GetServicePlanByGUID(guid string) (resources.ServicePlan, Warnings, error) {
	if guid == "" {
		return resources.ServicePlan{}, nil, ccerror.ServicePlanNotFound{}
	}

	var result resources.ServicePlan

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetServicePlanRequest,
		URIParams:    internal.Params{"service_plan_guid": guid},
		ResponseBody: &result,
	})

	return result, warnings, err
}

// GetServicePlans lists service plan with optional filters.
func (client *Client) GetServicePlans(query ...Query) ([]resources.ServicePlan, Warnings, error) {
	plans, _, warnings, err := client.getServicePlans(query...)
	return plans, warnings, err
}

func (client *Client) getServicePlans(query ...Query) ([]resources.ServicePlan, IncludedResources, Warnings, error) {
	var plans []resources.ServicePlan

	included, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetServicePlansRequest,
		Query:        query,
		ResponseBody: resources.ServicePlan{},
		AppendToList: func(item interface{}) error {
			plans = append(plans, item.(resources.ServicePlan))
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
	VisibilityType resources.ServicePlanVisibilityType
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

type ServiceOfferingWithPlans struct {
	// GUID is a unique service offering identifier.
	GUID string
	// Name is the name of the service offering.
	Name string
	// Description of the service offering
	Description string
	// ServiceBrokerName is the name of the service broker
	ServiceBrokerName string

	// List of service plans that this service offering provides
	Plans []resources.ServicePlan
}

func (client *Client) GetServicePlansWithOfferings(query ...Query) ([]ServiceOfferingWithPlans, Warnings, error) {
	query = append(query,
		Query{
			Key:    Include,
			Values: []string{"service_offering"},
		},
		Query{
			Key:    FieldsServiceOfferingServiceBroker,
			Values: []string{"name,guid"},
		},
	)

	plans, included, warnings, err := client.getServicePlans(query...)
	if err != nil {
		return nil, warnings, err
	}

	var offeringsWithPlans []ServiceOfferingWithPlans
	offeringGUIDLookup := make(map[string]int)

	indexOfOffering := func(serviceOfferingGUID string) int {
		if i, ok := offeringGUIDLookup[serviceOfferingGUID]; ok {
			return i
		}

		i := len(offeringsWithPlans)
		offeringGUIDLookup[serviceOfferingGUID] = i
		offeringsWithPlans = append(offeringsWithPlans, ServiceOfferingWithPlans{GUID: serviceOfferingGUID})

		return i
	}

	brokerNameLookup := lookuptable.NameFromGUID(included.ServiceBrokers)

	for _, p := range plans {
		i := indexOfOffering(p.ServiceOfferingGUID)
		offeringsWithPlans[i].Plans = append(offeringsWithPlans[i].Plans, p)
	}

	for _, o := range included.ServiceOfferings {
		i := indexOfOffering(o.GUID)
		offeringsWithPlans[i].Name = o.Name
		offeringsWithPlans[i].Description = o.Description
		offeringsWithPlans[i].ServiceBrokerName = brokerNameLookup[o.ServiceBrokerGUID]
	}

	return offeringsWithPlans, warnings, nil
}

func computeSpaceDetailsTable(included IncludedResources) map[string]planSpaceDetails {
	orgNameFromGUID := lookuptable.NameFromGUID(included.Organizations)

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
