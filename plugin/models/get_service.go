package plugin_models

type GetService_Model struct {
	Guid            string
	Name            string
	DashboardUrl    string
	IsUserProvided  bool
	ServiceOffering GetService_ServiceFields
	ServicePlan     GetService_ServicePlan
	LastOperation   GetService_LastOperation
}

type GetService_LastOperation struct {
	Type        string
	State       string
	Description string
	CreatedAt   string
	UpdatedAt   string
}

type GetService_ServicePlan struct {
	Name string
	Guid string
}

type GetService_ServiceFields struct {
	Name             string
	DocumentationUrl string
}
