package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"net/url"
	"strings"
)

type PaginatedServiceOfferingResources struct {
	Resources []ServiceOfferingResource
}

type ServiceOfferingResource struct {
	Metadata Metadata
	Entity   ServiceOfferingEntity
}

func (resource ServiceOfferingResource) ToFields() (fields cf.ServiceOfferingFields) {
	fields.Label = resource.Entity.Label
	fields.Version = resource.Entity.Version
	fields.Provider = resource.Entity.Provider
	fields.Description = resource.Entity.Description
	fields.Guid = resource.Metadata.Guid
	fields.DocumentationUrl = resource.Entity.DocumentationUrl
	return
}

func (resource ServiceOfferingResource) ToModel() (offering cf.ServiceOffering) {
	offering.ServiceOfferingFields = resource.ToFields()
	for _, p := range resource.Entity.ServicePlans {
		servicePlan := cf.ServicePlanFields{}
		servicePlan.Name = p.Entity.Name
		servicePlan.Guid = p.Metadata.Guid
		offering.Plans = append(offering.Plans, servicePlan)
	}
	return offering
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

func (resource ServicePlanResource) ToFields() (fields cf.ServicePlanFields) {
	fields.Guid = resource.Metadata.Guid
	fields.Name = resource.Entity.Name
	return
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

func (resource ServiceInstanceResource) ToFields() (fields cf.ServiceInstanceFields) {
	fields.Guid = resource.Metadata.Guid
	fields.Name = resource.Entity.Name
	return
}

func (resource ServiceInstanceResource) ToModel() (instance cf.ServiceInstance) {
	instance.ServiceInstanceFields = resource.ToFields()
	instance.ServicePlan = resource.Entity.ServicePlan.ToFields()
	instance.ServiceOffering = resource.Entity.ServicePlan.Entity.ServiceOffering.ToFields()

	instance.ServiceBindings = []cf.ServiceBindingFields{}
	for _, bindingResource := range resource.Entity.ServiceBindings {
		instance.ServiceBindings = append(instance.ServiceBindings, bindingResource.ToFields())
	}
	return
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

func (resource ServiceBindingResource) ToFields() (fields cf.ServiceBindingFields) {
	fields.Url = resource.Metadata.Url
	fields.Guid = resource.Metadata.Guid
	fields.AppGuid = resource.Entity.AppGuid
	return
}

type ServiceBindingEntity struct {
	AppGuid string `json:"app_guid"`
}

type ServiceRepository interface {
	GetServiceOfferings() (offerings cf.ServiceOfferings, apiResponse net.ApiResponse)
	FindInstanceByName(name string) (instance cf.ServiceInstance, apiResponse net.ApiResponse)
	CreateServiceInstance(name, planGuid string) (identicalAlreadyExists bool, apiResponse net.ApiResponse)
	RenameService(instance cf.ServiceInstance, newName string) (apiResponse net.ApiResponse)
	DeleteService(instance cf.ServiceInstance) (apiResponse net.ApiResponse)
}

type CloudControllerServiceRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
}

func NewCloudControllerServiceRepository(config *configuration.Configuration, gateway net.Gateway) (repo CloudControllerServiceRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerServiceRepository) GetServiceOfferings() (offerings cf.ServiceOfferings, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/services?inline-relations-depth=1", repo.config.Target)
	spaceGuid := repo.config.SpaceFields.Guid

	if spaceGuid != "" {
		path = fmt.Sprintf("%s/v2/spaces/%s/services?inline-relations-depth=1", repo.config.Target, spaceGuid)
	}

	resources := new(PaginatedServiceOfferingResources)
	apiResponse = repo.gateway.GetResource(path, repo.config.AccessToken, resources)
	if apiResponse.IsNotSuccessful() {
		return
	}

	for _, r := range resources.Resources {
		offerings = append(offerings, r.ToModel())
	}

	return
}

func (repo CloudControllerServiceRepository) FindInstanceByName(name string) (instance cf.ServiceInstance, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/spaces/%s/service_instances?return_user_provided_service_instances=true&q=%s&inline-relations-depth=2", repo.config.Target, repo.config.SpaceFields.Guid, url.QueryEscape("name:"+name))

	resources := new(PaginatedServiceInstanceResources)
	apiResponse = repo.gateway.GetResource(path, repo.config.AccessToken, resources)
	if apiResponse.IsNotSuccessful() {
		return
	}

	if len(resources.Resources) == 0 {
		apiResponse = net.NewNotFoundApiResponse("Service instance '%s' not found", name)
		return
	}

	resource := resources.Resources[0]
	instance = resource.ToModel()
	return
}

func (repo CloudControllerServiceRepository) CreateServiceInstance(name, planGuid string) (identicalAlreadyExists bool, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/service_instances", repo.config.Target)
	data := fmt.Sprintf(
		`{"name":"%s","service_plan_guid":"%s","space_guid":"%s"}`,
		name, planGuid, repo.config.SpaceFields.Guid,
	)

	apiResponse = repo.gateway.CreateResource(path, repo.config.AccessToken, strings.NewReader(data))

	if apiResponse.IsNotSuccessful() && apiResponse.ErrorCode == cf.SERVICE_INSTANCE_NAME_TAKEN {

		serviceInstance, findInstanceApiResponse := repo.FindInstanceByName(name)

		if !findInstanceApiResponse.IsNotSuccessful() &&
			serviceInstance.ServicePlan.Guid == planGuid {
			apiResponse = net.ApiResponse{}
			identicalAlreadyExists = true
			return
		}
	}
	return
}

func (repo CloudControllerServiceRepository) RenameService(instance cf.ServiceInstance, newName string) (apiResponse net.ApiResponse) {
	body := fmt.Sprintf(`{"name":"%s"}`, newName)
	path := fmt.Sprintf("%s/v2/service_instances/%s", repo.config.Target, instance.Guid)

	if instance.IsUserProvided() {
		path = fmt.Sprintf("%s/v2/user_provided_service_instances/%s", repo.config.Target, instance.Guid)
	}
	return repo.gateway.UpdateResource(path, repo.config.AccessToken, strings.NewReader(body))
}

func (repo CloudControllerServiceRepository) DeleteService(instance cf.ServiceInstance) (apiResponse net.ApiResponse) {
	if len(instance.ServiceBindings) > 0 {
		return net.NewApiResponseWithMessage("Cannot delete service instance, apps are still bound to it")
	}
	path := fmt.Sprintf("%s/v2/service_instances/%s", repo.config.Target, instance.Guid)
	return repo.gateway.DeleteResource(path, repo.config.AccessToken)
}
