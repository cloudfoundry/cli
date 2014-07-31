package models

type ServicePlanFields struct {
	Guid                string
	Name                string
	Free                bool
	Public              bool
	Active              bool
	ServiceOfferingGuid string
	OrgNames            []string
}

type ServicePlan struct {
	ServicePlanFields
	ServiceOffering ServiceOfferingFields
}
