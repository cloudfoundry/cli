package models

type ServicePlanFields struct {
	Guid   string
	Name   string
	Free   bool
	Public bool
	Description
	Active              bool
	ServiceOfferingGuid string
	OrgNames            []string
}

type ServicePlan struct {
	ServicePlanFields
	ServiceOffering ServiceOfferingFields
}
