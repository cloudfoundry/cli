package resources

import (
	"code.cloudfoundry.org/jsonry"
)

type ServicePlanVisibilityType string

const (
	ServicePlanVisibilityPublic       ServicePlanVisibilityType = "public"
	ServicePlanVisibilityOrganization ServicePlanVisibilityType = "organization"
	ServicePlanVisibilitySpace        ServicePlanVisibilityType = "space"
	ServicePlanVisibilityAdmin        ServicePlanVisibilityType = "admin"
)

type ServicePlanVisibilityDetail struct {
	// Name is the organization name
	Name string `json:"name,omitempty"`
	// GUID of the organization
	GUID string `json:"guid"`
}

func (s ServicePlanVisibilityDetail) OmitJSONry() bool {
	return s == ServicePlanVisibilityDetail{}
}

// ServicePlanVisibility represents a Cloud Controller V3 Service Plan Visibility.
type ServicePlanVisibility struct {
	// Type is one of 'public', 'organization', 'space' or 'admin'
	Type ServicePlanVisibilityType `json:"type"`

	// Organizations list of organizations for the service plan
	Organizations []ServicePlanVisibilityDetail `json:"organizations,omitempty"`

	// Space that the plan is visible in
	Space ServicePlanVisibilityDetail `json:"space"`
}

func (s *ServicePlanVisibility) UnmarshalJSON(data []byte) error {
	return jsonry.Unmarshal(data, s)
}

func (s ServicePlanVisibility) MarshalJSON() ([]byte, error) {
	return jsonry.Marshal(s)
}
