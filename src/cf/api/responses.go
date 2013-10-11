package api

import "time"

type Metadata struct {
	Guid string
	Url  string
}

type Entity struct {
	Name     string
	Host     string
	Label    string
	Provider string
	Password string `json:"auth_password"`
	Username string `json:"auth_username"`
	Url      string `json:"broker_url"`
}

type Resource struct {
	Metadata Metadata
	Entity   Entity
}

type PaginatedResources struct {
	Resources []Resource
}

type PaginatedApplicationResources struct {
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

type PaginatedEventResources struct {
	Resources []EventResource
}

type EventResource struct {
	Metadata Metadata
	Entity   EventEntity
}

type EventEntity struct {
	Timestamp       time.Time
	ExitDescription string `json:"exit_description"`
	ExitStatus      int    `json:"exit_status"`
	InstanceIndex   int    `json:"instance_index"`
}

type PaginatedRouteResources struct {
	Routes []RouteResource `json:"resources"`
}

type RouteResource struct {
	Metadata Metadata
	Entity   RouteEntity
}

type RouteEntity struct {
	Host   string
	Domain Resource
	Apps   []Resource
}

type PaginatedOrganizationResources struct {
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
	DiskQuota        uint64 `json:"disk_quota"`
	Urls             []string
	State            string
	ServiceNames     []string `json:"service_names"`
}

type RouteSummary struct {
	Guid   string
	Host   string
	Domain DomainSummary
}

type PaginatedDomainResources struct {
	Resources []DomainResource
}

type DomainResource struct {
	Metadata Metadata
	Entity   DomainEntity
}

type DomainEntity struct {
	Name                   string
	OwningOrganizationGuid string `json:"owning_organization_guid"`
	Spaces                 []SpaceResource
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
	Guid            string
	ServiceOffering ServiceOfferingSummary `json:"service"`
}

type ServiceOfferingSummary struct {
	Label    string
	Provider string
	Version  string
}

type PaginatedServiceOfferingResources struct {
	Resources []ServiceOfferingResource
}

type ServiceOfferingResource struct {
	Metadata Metadata
	Entity   ServiceOfferingEntity
}

type ServiceOfferingEntity struct {
	Label            string
	Version          string
	Description      string
	DocumentationUrl string `json:"documentation_url"`
	Provider         string
	ServicePlans     []ServicePlanResource `json:"service_plans"`
}

type ServicePlanResource struct {
	Metadata Metadata
	Entity   ServicePlanEntity
}

type ServicePlanEntity struct {
	Name            string
	ServiceOffering ServiceOfferingResource `json:"service"`
}

type PaginatedServiceInstanceResources struct {
	Resources []ServiceInstanceResource
}

type ServiceInstanceResource struct {
	Metadata Metadata
	Entity   ServiceInstanceEntity
}

type ServiceInstanceEntity struct {
	Name            string
	ServiceBindings []ServiceBindingResource `json:"service_bindings"`
	ServicePlan     ServicePlanResource      `json:"service_plan"`
}

type ServiceBindingResource struct {
	Metadata Metadata
	Entity   ServiceBindingEntity
}

type ServiceBindingEntity struct {
	AppGuid string `json:"app_guid"`
}

type PaginatedServiceBrokerResources struct {
	ServiceBrokers []ServiceBrokerResource `json:"resources"`
}

type ServiceBrokerResource struct {
	Metadata Metadata
	Entity   ServiceBrokerEntity
}

type ServiceBrokerEntity struct {
	Guid     string
	Name     string
	Password string `json:"auth_password"`
	Username string `json:"auth_username"`
	Url      string `json:"broker_url"`
}

type PaginatedSpaceResources struct {
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

type PaginatedStackResources struct {
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
