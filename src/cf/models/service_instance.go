package models

type ServiceInstanceFields struct {
	BasicFields
	SysLogDrainUrl   string
	ApplicationNames []string
	Params           map[string]string
}

type ServiceInstance struct {
	ServiceInstanceFields
	ServiceBindings []ServiceBindingFields
	ServicePlan     ServicePlanFields
	ServiceOffering ServiceOfferingFields
}

func (inst ServiceInstance) IsUserProvided() bool {
	return inst.ServicePlan.Guid == ""
}
