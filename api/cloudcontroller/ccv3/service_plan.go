package ccv3

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

// ServicePlan represents a Cloud Controller V3 Service Plan.
type ServicePlan struct {
	// GUID is a unique service plan identifier.
	GUID string
	// Name is the name of the service plan.
	Name string
	// VisibilityType can be "public", "admin", "organization" or "space"
	VisibilityType string
	// ServicePlanGUID is the GUID of the service offering
	ServiceOfferingGUID string

	Metadata *Metadata
}

func (sp *ServicePlan) UnmarshalJSON(data []byte) error {
	var response servicePlanResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return err
	}

	sp.GUID = response.GUID
	sp.Name = response.Name
	sp.VisibilityType = response.VisibilityType
	sp.ServiceOfferingGUID = response.Relationships.ServiceOffering.Data.GUID
	sp.Metadata = response.Metadata

	return nil
}

// servicePlanResponse represents a Cloud Controller V3 Service Plan (for reading)
type servicePlanResponse struct {
	// GUID is a unique service plan identifier.
	GUID string `json:"guid,required"`
	// Name is the name of the service plan.
	Name string `json:"name"`
	// VisibilityType can be "public", "admin", "organization" or "space"
	VisibilityType string `json:"visibility_type"`
	// Relationships represents related resources
	Relationships servicePlanRelationships `json:"relationships"`

	Metadata *Metadata
}

type servicePlanRelationships struct {
	ServiceOffering servicePlanRelationshipOffering `json:"service_offering"`
}

type servicePlanRelationshipOffering struct {
	Data servicePlanRelationshipOfferingData `json:"data"`
}

type servicePlanRelationshipOfferingData struct {
	GUID string `json:"guid"`
}

// GetServicePlans lists service plan with optional filters.
func (client *Client) GetServicePlans(query ...Query) ([]ServicePlan, Warnings, error) {
	var resources []ServicePlan

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetServicePlansRequest,
		Query:        query,
		ResponseBody: ServicePlan{},
		AppendToList: func(item interface{}) error {
			resources = append(resources, item.(ServicePlan))
			return nil
		},
	})

	return resources, warnings, err
}
