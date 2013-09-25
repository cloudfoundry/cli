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
	Name            string
	State           string
	Instances       int
	Memory          int
	Routes          []RouteResource
	EnvironmentJson map[string]string `json:"environment_json"`
}

type AppFile struct {
	Path string `json:"fn"`
	Sha1 string `json:"sha1"`
	Size int64  `json:"size"`
}

type RoutesResponse struct {
	Routes []RouteResource `json:"resources"`
}

type RouteResource struct {
	Metadata Metadata
	Entity   RouteEntity
}

type RouteEntity struct {
	Host   string
	Domain Resource
}

type OrganizationsApiResponse struct {
	Resources []OrganizationResource
}

type OrganizationResource struct {
	Metadata Metadata
	Entity   OrganizationEntity
}

type OrganizationEntity struct {
	Name    string
	Spaces  []Resource
	Domains []Resource
}

type ApplicationSummary struct {
	Guid             string
	Name             string
	Routes           []RouteSummary
	RunningInstances int `json:"running_instances"`
	Memory           uint64
	Instances        int
	Urls             []string
	State            string
	ServiceNames     []string `json:"service_names"`
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
	Guid             string
	Name             string
	Apps             []ApplicationSummary
	ServiceInstances []ServiceInstanceSummary `json:"services"`
}

type ServiceInstanceSummary struct {
	Name        string
	ServicePlan ServicePlanSummary `json:"service_plan"`
}

type ServicePlanSummary struct {
	Name            string
	ServiceOffering ServiceOfferingSummary `json:"service"`
}

type ServiceOfferingSummary struct {
	Label    string
	Provider string
	Version  string
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
	Version      string
	Description  string
	Provider     string
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

type SpaceApiResponse struct {
	Resources []SpaceResource
}

type SpaceResource struct {
	Metadata Metadata
	Entity   SpaceEntity
}

type SpaceEntity struct {
	Name             string
	Organization     Resource
	Applications     []Resource `json:"apps"`
	Domains          []Resource
	ServiceInstances []ServiceInstanceResource `json:"service_instances"`
}

type StackApiResponse struct {
	Resources []StackResource
}

type StackEntity struct {
	Name        string
	Description string
}

type StackResource struct {
	Metadata Metadata
	Entity   StackEntity
}
