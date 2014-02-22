package api

import (
	"cf"
	"cf/configuration"
	"cf/models"
	"cf/net"
	"fmt"
	"net/url"
	"strings"
)

type ServiceRepository interface {
	PurgeServiceOffering(offering models.ServiceOffering) net.ApiResponse
	FindServiceOfferingByLabelAndProvider(name, provider string) (offering models.ServiceOffering, apiResponse net.ApiResponse)
	GetAllServiceOfferings() (offerings models.ServiceOfferings, apiResponse net.ApiResponse)
	GetServiceOfferingsForSpace(spaceGuid string) (offerings models.ServiceOfferings, apiResponse net.ApiResponse)
	FindInstanceByName(name string) (instance models.ServiceInstance, apiResponse net.ApiResponse)
	CreateServiceInstance(name, planGuid string) (identicalAlreadyExists bool, apiResponse net.ApiResponse)
	RenameService(instance models.ServiceInstance, newName string) (apiResponse net.ApiResponse)
	DeleteService(instance models.ServiceInstance) (apiResponse net.ApiResponse)
	FindServicePlanByDescription(planDescription ServicePlanDescription) (planGuid string, apiResponse net.ApiResponse)
	GetServiceInstanceCountForServicePlan(v1PlanGuid string) (count int, apiResponse net.ApiResponse)
	MigrateServicePlanFromV1ToV2(v1PlanGuid, v2PlanGuid string) (changedCount int, apiResponse net.ApiResponse)
}

type CloudControllerServiceRepository struct {
	config  configuration.Reader
	gateway net.Gateway
}

func NewCloudControllerServiceRepository(config configuration.Reader, gateway net.Gateway) (repo CloudControllerServiceRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerServiceRepository) GetServiceOfferingsForSpace(spaceGuid string) (offerings models.ServiceOfferings, apiResponse net.ApiResponse) {
	return repo.getServiceOfferings(
		fmt.Sprintf("%s/v2/spaces/%s/services?inline-relations-depth=1", repo.config.ApiEndpoint(), spaceGuid),
	)
}

func (repo CloudControllerServiceRepository) GetAllServiceOfferings() (offerings models.ServiceOfferings, apiResponse net.ApiResponse) {
	return repo.getServiceOfferings(
		fmt.Sprintf("%s/v2/services?inline-relations-depth=1", repo.config.ApiEndpoint()),
	)
}

func (repo CloudControllerServiceRepository) getServiceOfferings(path string) (offerings models.ServiceOfferings, apiResponse net.ApiResponse) {
	resources := new(PaginatedServiceOfferingResources)
	apiResponse = repo.gateway.GetResource(path, repo.config.AccessToken(), resources)
	if apiResponse.IsNotSuccessful() {
		return
	}

	for _, r := range resources.Resources {
		offerings = append(offerings, r.ToModel())
	}

	return
}

func (repo CloudControllerServiceRepository) FindInstanceByName(name string) (instance models.ServiceInstance, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/spaces/%s/service_instances?return_user_provided_service_instances=true&q=%s&inline-relations-depth=2", repo.config.ApiEndpoint(), repo.config.SpaceFields().Guid, url.QueryEscape("name:"+name))

	resources := new(PaginatedServiceInstanceResources)
	apiResponse = repo.gateway.GetResource(path, repo.config.AccessToken(), resources)
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
	path := fmt.Sprintf("%s/v2/service_instances", repo.config.ApiEndpoint())
	data := fmt.Sprintf(
		`{"name":"%s","service_plan_guid":"%s","space_guid":"%s", "async": true}`,
		name, planGuid, repo.config.SpaceFields().Guid,
	)

	apiResponse = repo.gateway.CreateResource(path, repo.config.AccessToken(), strings.NewReader(data))

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

func (repo CloudControllerServiceRepository) RenameService(instance models.ServiceInstance, newName string) (apiResponse net.ApiResponse) {
	body := fmt.Sprintf(`{"name":"%s"}`, newName)
	path := fmt.Sprintf("%s/v2/service_instances/%s", repo.config.ApiEndpoint(), instance.Guid)

	if instance.IsUserProvided() {
		path = fmt.Sprintf("%s/v2/user_provided_service_instances/%s", repo.config.ApiEndpoint(), instance.Guid)
	}
	return repo.gateway.UpdateResource(path, repo.config.AccessToken(), strings.NewReader(body))
}

func (repo CloudControllerServiceRepository) DeleteService(instance models.ServiceInstance) (apiResponse net.ApiResponse) {
	if len(instance.ServiceBindings) > 0 {
		return net.NewApiResponseWithMessage("Cannot delete service instance, apps are still bound to it")
	}
	path := fmt.Sprintf("%s/v2/service_instances/%s", repo.config.ApiEndpoint(), instance.Guid)
	return repo.gateway.DeleteResource(path, repo.config.AccessToken())
}

func (repo CloudControllerServiceRepository) PurgeServiceOffering(offering models.ServiceOffering) net.ApiResponse {
	url := fmt.Sprintf("%s/v2/services/%s?purge=true", repo.config.ApiEndpoint(), offering.Guid)
	return repo.gateway.DeleteResource(url, repo.config.AccessToken())
}

func (repo CloudControllerServiceRepository) FindServiceOfferingByLabelAndProvider(label, provider string) (offering models.ServiceOffering, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/services?q=%s", repo.config.ApiEndpoint(), url.QueryEscape("label:"+label+";provider:"+provider))

	resources := new(PaginatedServiceOfferingResources)
	apiResponse = repo.gateway.GetResource(path, repo.config.AccessToken(), resources)

	if apiResponse.IsError() {
		return
	} else if len(resources.Resources) == 0 {
		apiResponse = net.NewNotFoundApiResponse("Service offering not found")
	} else {
		offering = resources.Resources[0].ToModel()
	}
	return
}

func (repo CloudControllerServiceRepository) FindServicePlanByDescription(planDescription ServicePlanDescription) (planGuid string, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/services?inline-relations-depth=1&q=%s",
		repo.config.ApiEndpoint(),
		url.QueryEscape("label:"+planDescription.ServiceName+";provider:"+planDescription.ServiceProvider))

	response := new(PaginatedServiceOfferingResources)
	apiResponse = repo.gateway.GetResource(path, repo.config.AccessToken(), response)
	if apiResponse.IsNotSuccessful() {
		return
	}

	for _, serviceOfferingResource := range response.Resources {
		for _, servicePlanResource := range serviceOfferingResource.Entity.ServicePlans {
			if servicePlanResource.Entity.Name == planDescription.ServicePlanName {
				planGuid = servicePlanResource.Metadata.Guid
				return
			}
		}
	}

	apiResponse = net.NewNotFoundApiResponse("Plan %s cannot be found", planDescription)

	return
}

func (repo CloudControllerServiceRepository) MigrateServicePlanFromV1ToV2(v1PlanGuid, v2PlanGuid string) (changedCount int, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/service_plans/%s/service_instances", repo.config.ApiEndpoint(), v1PlanGuid)
	body := strings.NewReader(fmt.Sprintf(`{"service_plan_guid":"%s"}`, v2PlanGuid))
	response := new(ServiceMigrateV1ToV2Response)

	apiResponse = repo.gateway.UpdateResourceForResponse(path, repo.config.AccessToken(), body, response)
	if apiResponse.IsNotSuccessful() {
		return
	}

	changedCount = response.ChangedCount
	return
}

func (repo CloudControllerServiceRepository) GetServiceInstanceCountForServicePlan(v1PlanGuid string) (count int, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/service_plans/%s/service_instances?results-per-page=1", repo.config.ApiEndpoint(), v1PlanGuid)
	response := new(PaginatedServiceInstanceResources)
	apiResponse = repo.gateway.GetResource(path, repo.config.AccessToken(), response)
	count = response.TotalResults
	return
}
