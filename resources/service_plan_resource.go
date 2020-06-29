package resources

import "code.cloudfoundry.org/jsonry"

type ServicePlanCost struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	Unit     string  `json:"unit"`
}

type ServicePlan struct {
	// GUID is a unique service plan identifier.
	GUID string `json:"guid"`
	// Name is the name of the service plan.
	Name string `json:"name"`
	// Description of the Service Plan.
	Description string `json:"description"`
	// Whether the Service Plan is available
	Available bool `json:"available"`
	// VisibilityType can be "public", "admin", "organization" or "space"
	VisibilityType ServicePlanVisibilityType `json:"visibility_type"`
	// Free shows whether or not the Service Plan is free of charge.
	Free bool `json:"free"`
	// Cost shows the cost of a paid service plan
	Costs []ServicePlanCost `json:"costs"`
	// ServicePlanGUID is the GUID of the service offering
	ServiceOfferingGUID string `jsonry:"relationships.service_offering.data.guid"`
	// SpaceGUID is the space that a plan from a space-scoped broker relates to
	SpaceGUID string `jsonry:"relationships.space.data.guid"`

	Metadata *Metadata `json:"metadata"`
}

func (p *ServicePlan) UnmarshalJSON(data []byte) error {
	return jsonry.Unmarshal(data, p)
}
