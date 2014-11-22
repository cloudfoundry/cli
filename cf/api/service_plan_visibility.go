package api

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
)

type ServicePlanVisibilityRepository interface {
	Create(string, string) error
	List() ([]models.ServicePlanVisibilityFields, error)
	Delete(string) error
	Search(map[string]string) ([]models.ServicePlanVisibilityFields, error)
}

type CloudControllerServicePlanVisibilityRepository struct {
	config  core_config.Reader
	gateway net.Gateway
}

func NewCloudControllerServicePlanVisibilityRepository(config core_config.Reader, gateway net.Gateway) CloudControllerServicePlanVisibilityRepository {
	return CloudControllerServicePlanVisibilityRepository{
		config:  config,
		gateway: gateway,
	}
}

func (repo CloudControllerServicePlanVisibilityRepository) Create(serviceGuid, orgGuid string) error {
	url := "/v2/service_plan_visibilities"
	data := fmt.Sprintf(`{"service_plan_guid":"%s", "organization_guid":"%s"}`, serviceGuid, orgGuid)
	return repo.gateway.CreateResource(repo.config.ApiEndpoint(), url, strings.NewReader(data))
}

func (repo CloudControllerServicePlanVisibilityRepository) List() (visibilities []models.ServicePlanVisibilityFields, err error) {
	err = repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		"/v2/service_plan_visibilities",
		resources.ServicePlanVisibilityResource{},
		func(resource interface{}) bool {
			if spv, ok := resource.(resources.ServicePlanVisibilityResource); ok {
				visibilities = append(visibilities, spv.ToFields())
			}
			return true
		})
	return
}

func (repo CloudControllerServicePlanVisibilityRepository) Delete(servicePlanGuid string) error {
	path := fmt.Sprintf("/v2/service_plan_visibilities/%s", servicePlanGuid)
	return repo.gateway.DeleteResource(repo.config.ApiEndpoint(), path)
}

func (repo CloudControllerServicePlanVisibilityRepository) Search(queryParams map[string]string) ([]models.ServicePlanVisibilityFields, error) {
	var visibilities []models.ServicePlanVisibilityFields
	err := repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		combineQueryParametersWithUri("/v2/service_plan_visibilities", queryParams),
		resources.ServicePlanVisibilityResource{},
		func(resource interface{}) bool {
			if sp, ok := resource.(resources.ServicePlanVisibilityResource); ok {
				visibilities = append(visibilities, sp.ToFields())
			}
			return true
		})
	return visibilities, err
}
