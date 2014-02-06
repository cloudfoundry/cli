package models

type ServicePlanFields struct {
	BasicFields
}

type ServicePlan struct {
	ServicePlanFields
	ServiceOffering ServiceOfferingFields
}
