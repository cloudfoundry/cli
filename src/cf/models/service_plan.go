package models

type ServicePlanFields struct {
	Guid string
	Name string
}

type ServicePlan struct {
	ServicePlanFields
	ServiceOffering ServiceOfferingFields
}
