package models

type ServicePlanVisibilityFields struct {
	Guid             string `json:"guid"`
	ServicePlanGuid  string `json:"service_plan_guid"`
	OrganizationGuid string `json:"organization_guid"`
}
