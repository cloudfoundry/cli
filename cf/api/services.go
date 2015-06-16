package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
)

type ServiceRepository interface {
	PurgeServiceOffering(offering models.ServiceOffering) error
	GetServiceOfferingByGuid(serviceGuid string) (offering models.ServiceOffering, apiErr error)
	FindServiceOfferingsByLabel(name string) (offering models.ServiceOfferings, apiErr error)
	FindServiceOfferingByLabelAndProvider(name, provider string) (offering models.ServiceOffering, apiErr error)

	FindServiceOfferingsForSpaceByLabel(spaceGuid, name string) (offering models.ServiceOfferings, apiErr error)

	GetAllServiceOfferings() (offerings models.ServiceOfferings, apiErr error)
	GetServiceOfferingsForSpace(spaceGuid string) (offerings models.ServiceOfferings, apiErr error)
	FindInstanceByName(name string) (instance models.ServiceInstance, apiErr error)
	CreateServiceInstance(name, planGuid string, params map[string]interface{}, tags []string) (apiErr error)
	UpdateServiceInstance(instanceGuid, planGuid string, params map[string]interface{}, tags []string) (apiErr error)
	RenameService(instance models.ServiceInstance, newName string) (apiErr error)
	DeleteService(instance models.ServiceInstance) (apiErr error)
	FindServicePlanByDescription(planDescription resources.ServicePlanDescription) (planGuid string, apiErr error)
	ListServicesFromBroker(brokerGuid string) (services []models.ServiceOffering, err error)
	GetServiceInstanceCountForServicePlan(v1PlanGuid string) (count int, apiErr error)
	MigrateServicePlanFromV1ToV2(v1PlanGuid, v2PlanGuid string) (changedCount int, apiErr error)
}

type CloudControllerServiceRepository struct {
	config  core_config.Reader
	gateway net.Gateway
}

func NewCloudControllerServiceRepository(config core_config.Reader, gateway net.Gateway) (repo CloudControllerServiceRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerServiceRepository) GetServiceOfferingByGuid(serviceGuid string) (models.ServiceOffering, error) {
	offering := new(resources.ServiceOfferingResource)
	apiErr := repo.gateway.GetResource(repo.config.ApiEndpoint()+fmt.Sprintf("/v2/services/%s", serviceGuid), offering)
	serviceOffering := offering.ToFields()
	return models.ServiceOffering{ServiceOfferingFields: serviceOffering}, apiErr
}

func (repo CloudControllerServiceRepository) GetServiceOfferingsForSpace(spaceGuid string) (models.ServiceOfferings, error) {
	return repo.getServiceOfferings(fmt.Sprintf("/v2/spaces/%s/services", spaceGuid))
}

func (repo CloudControllerServiceRepository) FindServiceOfferingsForSpaceByLabel(spaceGuid, name string) (offerings models.ServiceOfferings, err error) {
	offerings, err = repo.getServiceOfferings(fmt.Sprintf("/v2/spaces/%s/services?q=%s", spaceGuid, url.QueryEscape("label:"+name)))

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

func (repo CloudControllerServiceRepository) GetAllServiceOfferings() (models.ServiceOfferings, error) {
	return repo.getServiceOfferings("/v2/services")
}

func (repo CloudControllerServiceRepository) getServiceOfferings(path string) ([]models.ServiceOffering, error) {
	var offerings []models.ServiceOffering
	apiErr := repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		path,
		resources.ServiceOfferingResource{},
		func(resource interface{}) bool {
			if so, ok := resource.(resources.ServiceOfferingResource); ok {
				offerings = append(offerings, so.ToModel())
			}
			return true
		})

	return offerings, apiErr
}

func (repo CloudControllerServiceRepository) FindInstanceByName(name string) (instance models.ServiceInstance, apiErr error) {
	path := fmt.Sprintf("%s/v2/spaces/%s/service_instances?return_user_provided_service_instances=true&q=%s&inline-relations-depth=1", repo.config.ApiEndpoint(), repo.config.SpaceFields().Guid, url.QueryEscape("name:"+name))

	responseJSON := new(resources.PaginatedServiceInstanceResources)
	apiErr = repo.gateway.GetResource(path, responseJSON)
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
		apiErr = repo.gateway.GetResource(path, resource)
		instance.ServiceOffering = resource.ToFields()
	}

	return
}

func (repo CloudControllerServiceRepository) CreateServiceInstance(name, planGuid string, params map[string]interface{}, tags []string) (err error) {
	path := "/v2/service_instances?accepts_incomplete=true"
	request := models.ServiceInstanceCreateRequest{
		Name:      name,
		PlanGuid:  planGuid,
		SpaceGuid: repo.config.SpaceFields().Guid,
		Params:    params,
		Tags:      tags,
	}

	jsonBytes, err := json.Marshal(request)
	if err != nil {
		return err
	}

	err = repo.gateway.CreateResource(repo.config.ApiEndpoint(), path, bytes.NewReader(jsonBytes))

	if httpErr, ok := err.(errors.HttpError); ok && httpErr.ErrorCode() == errors.SERVICE_INSTANCE_NAME_TAKEN {
		_, findInstanceErr := repo.FindInstanceByName(name)

		if nil == findInstanceErr {
			return errors.NewModelAlreadyExistsError("Service", name)
		}
	}

	return
}

func (repo CloudControllerServiceRepository) UpdateServiceInstance(instanceGuid, planGuid string, params map[string]interface{}, tags []string) (err error) {
	path := fmt.Sprintf("/v2/service_instances/%s?accepts_incomplete=true", instanceGuid)
	request := models.ServiceInstanceUpdateRequest{
		PlanGuid: planGuid,
		Params:   params,
		Tags:     tags,
	}

	jsonBytes, err := json.Marshal(request)
	if err != nil {
		return err
	}

	err = repo.gateway.UpdateResource(repo.config.ApiEndpoint(), path, bytes.NewReader(jsonBytes))

	return
}

func (repo CloudControllerServiceRepository) RenameService(instance models.ServiceInstance, newName string) (apiErr error) {
	body := fmt.Sprintf(`{"name":"%s"}`, newName)
	path := fmt.Sprintf("/v2/service_instances/%s?accepts_incomplete=true", instance.Guid)

	if instance.IsUserProvided() {
		path = fmt.Sprintf("/v2/user_provided_service_instances/%s", instance.Guid)
	}
	return repo.gateway.UpdateResource(repo.config.ApiEndpoint(), path, strings.NewReader(body))
}

func (repo CloudControllerServiceRepository) DeleteService(instance models.ServiceInstance) (apiErr error) {
	if len(instance.ServiceBindings) > 0 || len(instance.ServiceKeys) > 0 {
		return errors.NewServiceAssociationError()
	}
	path := fmt.Sprintf("/v2/service_instances/%s?%s", instance.Guid, "accepts_incomplete=true")
	return repo.gateway.DeleteResource(repo.config.ApiEndpoint(), path)
}

func (repo CloudControllerServiceRepository) PurgeServiceOffering(offering models.ServiceOffering) error {
	url := fmt.Sprintf("/v2/services/%s?purge=true", offering.Guid)
	return repo.gateway.DeleteResource(repo.config.ApiEndpoint(), url)
}

func (repo CloudControllerServiceRepository) FindServiceOfferingsByLabel(label string) (models.ServiceOfferings, error) {
	path := fmt.Sprintf("/v2/services?q=%s", url.QueryEscape("label:"+label))
	offerings, apiErr := repo.getServiceOfferings(path)

	if apiErr != nil {
		return models.ServiceOfferings{}, apiErr
	} else if len(offerings) == 0 {
		apiErr = errors.NewModelNotFoundError("Service offering", label)
		return models.ServiceOfferings{}, apiErr
	}

	return offerings, apiErr
}

func (repo CloudControllerServiceRepository) FindServiceOfferingByLabelAndProvider(label, provider string) (models.ServiceOffering, error) {
	path := fmt.Sprintf("/v2/services?q=%s", url.QueryEscape("label:"+label+";provider:"+provider))
	offerings, apiErr := repo.getServiceOfferings(path)

	if apiErr != nil {
		return models.ServiceOffering{}, apiErr
	} else if len(offerings) == 0 {
		apiErr = errors.NewModelNotFoundError("Service offering", label+" "+provider)
		return models.ServiceOffering{}, apiErr
	}

	return offerings[0], apiErr
}

func (repo CloudControllerServiceRepository) FindServicePlanByDescription(planDescription resources.ServicePlanDescription) (string, error) {
	path := fmt.Sprintf("/v2/services?inline-relations-depth=1&q=%s",
		url.QueryEscape("label:"+planDescription.ServiceLabel+";provider:"+planDescription.ServiceProvider))

	var planGuid string
	offerings, apiErr := repo.getServiceOfferings(path)
	if apiErr != nil {
		return planGuid, apiErr
	}

	for _, serviceOfferingResource := range offerings {
		for _, servicePlanResource := range serviceOfferingResource.Plans {
			if servicePlanResource.Name == planDescription.ServicePlanName {
				planGuid := servicePlanResource.Guid
				return planGuid, apiErr
			}
		}
	}

	apiErr = errors.NewModelNotFoundError("Plan", planDescription.String())

	return planGuid, apiErr
}

func (repo CloudControllerServiceRepository) ListServicesFromBroker(brokerGuid string) (offerings []models.ServiceOffering, err error) {
	err = repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		fmt.Sprintf("/v2/services?q=%s", url.QueryEscape("service_broker_guid:"+brokerGuid)),
		resources.ServiceOfferingResource{},
		func(resource interface{}) bool {
			if offering, ok := resource.(resources.ServiceOfferingResource); ok {
				offerings = append(offerings, offering.ToModel())
			}
			return true
		})
	return
}

func (repo CloudControllerServiceRepository) MigrateServicePlanFromV1ToV2(v1PlanGuid, v2PlanGuid string) (changedCount int, apiErr error) {
	path := fmt.Sprintf("/v2/service_plans/%s/service_instances", v1PlanGuid)
	body := strings.NewReader(fmt.Sprintf(`{"service_plan_guid":"%s"}`, v2PlanGuid))
	response := new(resources.ServiceMigrateV1ToV2Response)

	apiErr = repo.gateway.UpdateResource(repo.config.ApiEndpoint(), path, body, response)
	if apiErr != nil {
		return
	}

	changedCount = response.ChangedCount
	return
}

func (repo CloudControllerServiceRepository) GetServiceInstanceCountForServicePlan(v1PlanGuid string) (count int, apiErr error) {
	path := fmt.Sprintf("%s/v2/service_plans/%s/service_instances?results-per-page=1", repo.config.ApiEndpoint(), v1PlanGuid)
	response := new(resources.PaginatedServiceInstanceResources)
	apiErr = repo.gateway.GetResource(path, response)
	count = response.TotalResults
	return
}
