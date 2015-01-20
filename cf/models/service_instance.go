package models

type ServiceInstanceFields struct {
	Guid             string
	Name             string
	State            string
	StateDescription string
	SysLogDrainUrl   string
	ApplicationNames []string
	Params           map[string]interface{}
	DashboardUrl     string
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
