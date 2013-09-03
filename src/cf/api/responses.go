package api

type Metadata struct {
	Guid string
	Url  string
}

type Entity struct {
	Name string
	Host string
}

type Resource struct {
	Metadata Metadata
	Entity   Entity
}

type ApiResponse struct {
	Resources []Resource
}

type ApplicationsApiResponse struct {
	Resources []ApplicationResource
}

type ApplicationResource struct {
	Metadata Metadata
	Entity   ApplicationEntity
}

type ApplicationEntity struct {
	Name      string
	State     string
	Instances int
	Memory    int
	Routes    []RouteResource
}

type RouteResource struct {
	Metadata Metadata
	Entity   RouteEntity
}

type RouteEntity struct {
	Host   string
	Domain Resource
}

type ApplicationSummary struct {
	Guid             string
	Name             string
	Routes           []RouteSummary
	RunningInstances int `json:"running_instances"`
	Memory           int
	Instances        int
	Urls             []string
	State            string
}

type RouteSummary struct {
	Guid   string
	Host   string
	Domain DomainSummary
}

type DomainSummary struct {
	Guid string
	Name string
}

type SpaceSummary struct {
	Guid string
	Name string
	Apps []ApplicationSummary
}

type ServiceOfferingsApiResponse struct {
	Resources []ServiceOfferingResource
}

type ServiceOfferingResource struct {
	Metadata Metadata
	Entity   ServiceOfferingEntity
}

type ServiceOfferingEntity struct {
	Label        string
	ServicePlans []ServicePlanResource `json:"service_plans"`
}

type ServicePlanResource struct {
	Metadata Metadata
	Entity   Entity
}

type ServiceInstancesApiResponse struct {
	Resources []ServiceInstanceResource
}

type ServiceInstanceResource struct {
	Metadata Metadata
	Entity   ServiceInstanceEntity
}

type ServiceInstanceEntity struct {
	Name            string
	ServiceBindings []ServiceBindingResource `json:"service_bindings"`
}

type ServiceBindingResource struct {
	Metadata Metadata
	Entity   ServiceBindingEntity
}

type ServiceBindingEntity struct {
	AppGuid string `json:"app_guid"`
}
