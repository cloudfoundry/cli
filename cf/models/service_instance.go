package models

type LastOperationFields struct {
	Type        string
	State       string
	Description string
}

type ServiceInstanceFields struct {
	Guid             string
	Name             string
	LastOperation    LastOperationFields
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
