package plugin_models

type LastOperationFields struct {
	Type        string
	State       string
	Description string
}

type ServicePlanFields struct {
	Guid string
	Name string
}

type ServiceFields struct {
	Name string
}

type ServiceInstance struct {
	Guid             string
	Name             string
	ServicePlan      ServicePlanFields
	Service          ServiceFields
	LastOperation    LastOperationFields
	ApplicationNames []string
	IsUserProvided   bool
}
