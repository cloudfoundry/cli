package models

type LastOperationFields struct {
	Type        string
	State       string
	Description string
}

type ServiceInstanceRequest struct {
	Name      string                 `json:"name"`
	SpaceGuid string                 `json:"space_guid"`
	PlanGuid  string                 `json:"service_plan_guid"`
	Params    map[string]interface{} `json:"parameters,omitempty"`
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
