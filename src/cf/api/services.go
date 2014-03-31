package api

import (
	"cf/api/resources"
	"cf/configuration"
	"cf/errors"
	"cf/models"
	"cf/net"
	"fmt"
	"net/url"
	"strings"
)

type ServiceRepository interface {
	PurgeServiceOffering(offering models.ServiceOffering) error
	FindServiceOfferingByLabelAndProvider(name, provider string) (offering models.ServiceOffering, apiErr error)
	FindServiceOfferingsForSpaceByLabel(spaceGuid, name string) (offering models.ServiceOfferings, apiErr error)
	GetAllServiceOfferings() (offerings models.ServiceOfferings, apiErr error)
	GetServiceOfferingsForSpace(spaceGuid string) (offerings models.ServiceOfferings, apiErr error)
	FindInstanceByName(name string) (instance models.ServiceInstance, apiErr error)
	CreateServiceInstance(name, planGuid string) (apiErr error)
	RenameService(instance models.ServiceInstance, newName string) (apiErr error)
	DeleteService(instance models.ServiceInstance) (apiErr error)
	FindServicePlanByDescription(planDescription resources.ServicePlanDescription) (planGuid string, apiErr error)
	GetServiceInstanceCountForServicePlan(v1PlanGuid string) (count int, apiErr error)
	MigrateServicePlanFromV1ToV2(v1PlanGuid, v2PlanGuid string) (changedCount int, apiErr error)
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

func (repo CloudControllerServiceRepository) FindServiceOfferingsForSpaceByLabel(spaceGuid, name string) (offerings models.ServiceOfferings, err error) {
	offerings, err = repo.getServiceOfferings(
		fmt.Sprintf("/v2/spaces/%s/services?q=%s&inline-relations-depth=1", spaceGuid, url.QueryEscape("label:"+name)))

	if httpErr, ok := err.(errors.HttpError); ok && httpErr.ErrorCode() == errors.BAD_QUERY_PARAM {
		offerings, err = repo.findServiceOfferingsByPaginating(spaceGuid, name)
	}

	if err == nil && len(offerings) == 0 {
		err = errors.NewModelNotFoundError("Service offering", name)
	}

	return
}

func (repo CloudControllerServiceRepository) findServiceOfferingsByPaginating(spaceGuid, label string) (offerings models.ServiceOfferings, apiErr error) {
	offerings, apiErr = repo.GetServiceOfferingsForSpace(spaceGuid)
	if apiErr != nil {
		return
	}

	matchingOffering := models.ServiceOfferings{}

	for _, offering := range offerings {
		if offering.Label == label {
			matchingOffering = append(matchingOffering, offering)
		}
	}
	return matchingOffering, nil
}

func (repo CloudControllerServiceRepository) GetServiceOfferingsForSpace(spaceGuid string) (offerings models.ServiceOfferings, apiErr error) {
	return repo.getServiceOfferings(
		fmt.Sprintf("/v2/spaces/%s/services?inline-relations-depth=1", spaceGuid))
}

func (repo CloudControllerServiceRepository) GetAllServiceOfferings() (offerings models.ServiceOfferings, apiErr error) {
	return repo.getServiceOfferings("/v2/services?inline-relations-depth=1")
}

func (repo CloudControllerServiceRepository) getServiceOfferings(path string) (offerings models.ServiceOfferings, apiErr error) {
	apiErr = repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		repo.config.AccessToken(),
		path,
		resources.ServiceOfferingResource{},
		func(resource interface{}) bool {
			if so, ok := resource.(resources.ServiceOfferingResource); ok {
				offerings = append(offerings, so.ToModel())
			}
			return true
		})
	if apiErr != nil {
		return
	}

	return
}

func (repo CloudControllerServiceRepository) FindInstanceByName(name string) (instance models.ServiceInstance, apiErr error) {
	path := fmt.Sprintf("%s/v2/spaces/%s/service_instances?return_user_provided_service_instances=true&q=%s&inline-relations-depth=1", repo.config.ApiEndpoint(), repo.config.SpaceFields().Guid, url.QueryEscape("name:"+name))

	responseJSON := new(resources.PaginatedServiceInstanceResources)
	apiErr = repo.gateway.GetResource(path, repo.config.AccessToken(), responseJSON)
	if apiErr != nil {
		return
	}

	if len(responseJSON.Resources) == 0 {
		apiErr = errors.NewModelNotFoundError("Service instance", name)
		return
	}

	instanceResource := responseJSON.Resources[0]
	instance = instanceResource.ToModel()

	if instanceResource.Entity.ServicePlan.Metadata.Guid != "" {
		resource := &resources.ServiceOfferingResource{}
		path = fmt.Sprintf("%s/v2/services/%s", repo.config.ApiEndpoint(), instanceResource.Entity.ServicePlan.Entity.ServiceOfferingGuid)
		apiErr = repo.gateway.GetResource(path, repo.config.AccessToken(), resource)
		instance.ServiceOffering = resource.ToFields()
	}

	return
}

func (repo CloudControllerServiceRepository) CreateServiceInstance(name, planGuid string) (err error) {
	path := fmt.Sprintf("%s/v2/service_instances", repo.config.ApiEndpoint())
	data := fmt.Sprintf(
		`{"name":"%s","service_plan_guid":"%s","space_guid":"%s", "async": true}`,
		name, planGuid, repo.config.SpaceFields().Guid,
	)

	err = repo.gateway.CreateResource(path, repo.config.AccessToken(), strings.NewReader(data))

	if httpErr, ok := err.(errors.HttpError); ok && httpErr.ErrorCode() == errors.SERVICE_INSTANCE_NAME_TAKEN {
		serviceInstance, findInstanceErr := repo.FindInstanceByName(name)

		if findInstanceErr == nil && serviceInstance.ServicePlan.Guid == planGuid {
			return errors.NewServiceInstanceAlreadyExistsError(name)
		}
	}

	return
}

func (repo CloudControllerServiceRepository) RenameService(instance models.ServiceInstance, newName string) (apiErr error) {
	body := fmt.Sprintf(`{"name":"%s"}`, newName)
	path := fmt.Sprintf("%s/v2/service_instances/%s", repo.config.ApiEndpoint(), instance.Guid)

	if instance.IsUserProvided() {
		path = fmt.Sprintf("%s/v2/user_provided_service_instances/%s", repo.config.ApiEndpoint(), instance.Guid)
	}
	return repo.gateway.UpdateResource(path, repo.config.AccessToken(), strings.NewReader(body))
}

func (repo CloudControllerServiceRepository) DeleteService(instance models.ServiceInstance) (apiErr error) {
	if len(instance.ServiceBindings) > 0 {
		return errors.New("Cannot delete service instance, apps are still bound to it")
	}
	path := fmt.Sprintf("%s/v2/service_instances/%s", repo.config.ApiEndpoint(), instance.Guid)
	return repo.gateway.DeleteResource(path, repo.config.AccessToken())
}

func (repo CloudControllerServiceRepository) PurgeServiceOffering(offering models.ServiceOffering) error {
	url := fmt.Sprintf("%s/v2/services/%s?purge=true", repo.config.ApiEndpoint(), offering.Guid)
	return repo.gateway.DeleteResource(url, repo.config.AccessToken())
}

func (repo CloudControllerServiceRepository) FindServiceOfferingByLabelAndProvider(label, provider string) (offering models.ServiceOffering, apiErr error) {
	path := fmt.Sprintf("%s/v2/services?q=%s", repo.config.ApiEndpoint(), url.QueryEscape("label:"+label+";provider:"+provider))

	resources := new(resources.PaginatedServiceOfferingResources)
	apiErr = repo.gateway.GetResource(path, repo.config.AccessToken(), resources)

	if apiErr != nil {
		return
	} else if len(resources.Resources) == 0 {
		apiErr = errors.NewModelNotFoundError("Service offering", label+" "+provider)
	} else {
		offering = resources.Resources[0].ToModel()
	}
	return
}

func (repo CloudControllerServiceRepository) FindServicePlanByDescription(planDescription resources.ServicePlanDescription) (planGuid string, apiErr error) {
	path := fmt.Sprintf("%s/v2/services?inline-relations-depth=1&q=%s",
		repo.config.ApiEndpoint(),
		url.QueryEscape("label:"+planDescription.ServiceLabel+";provider:"+planDescription.ServiceProvider))

	response := new(resources.PaginatedServiceOfferingResources)
	apiErr = repo.gateway.GetResource(path, repo.config.AccessToken(), response)
	if apiErr != nil {
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

	apiErr = errors.NewModelNotFoundError("Plan", planDescription.String())

	return
}

func (repo CloudControllerServiceRepository) MigrateServicePlanFromV1ToV2(v1PlanGuid, v2PlanGuid string) (changedCount int, apiErr error) {
	path := fmt.Sprintf("%s/v2/service_plans/%s/service_instances", repo.config.ApiEndpoint(), v1PlanGuid)
	body := strings.NewReader(fmt.Sprintf(`{"service_plan_guid":"%s"}`, v2PlanGuid))
	response := new(resources.ServiceMigrateV1ToV2Response)

	apiErr = repo.gateway.UpdateResourceForResponse(path, repo.config.AccessToken(), body, response)
	if apiErr != nil {
		return
	}

	changedCount = response.ChangedCount
	return
}

func (repo CloudControllerServiceRepository) GetServiceInstanceCountForServicePlan(v1PlanGuid string) (count int, apiErr error) {
	path := fmt.Sprintf("%s/v2/service_plans/%s/service_instances?results-per-page=1", repo.config.ApiEndpoint(), v1PlanGuid)
	response := new(resources.PaginatedServiceInstanceResources)
	apiErr = repo.gateway.GetResource(path, repo.config.AccessToken(), response)
	count = response.TotalResults
	return
}
