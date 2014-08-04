package api

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
)

type ServicePlanVisibilityRepository interface {
	Create(string, string) error
	List() ([]models.ServicePlanVisibilityFields, error)
	Delete(string) error
}

type CloudControllerServicePlanVisibilityRepository struct {
	config  configuration.Reader
	gateway net.Gateway
}

func NewCloudControllerServicePlanVisibilityRepository(config configuration.Reader, gateway net.Gateway) CloudControllerServicePlanVisibilityRepository {
	return CloudControllerServicePlanVisibilityRepository{
		config:  config,
		gateway: gateway,
	}
}

func (repo CloudControllerServicePlanVisibilityRepository) Create(serviceGuid, orgGuid string) error {
	url := repo.config.ApiEndpoint() + "/v2/service_plan_visibilities"
	data := fmt.Sprintf(`{"service_plan_guid":"%s", "organization_guid":"%s"}`, serviceGuid, orgGuid)
	return repo.gateway.CreateResource(url, strings.NewReader(data))
}

func (repo CloudControllerServicePlanVisibilityRepository) List() (plans []models.ServicePlanVisibilityFields, err error) {
	err = repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		"/v2/service_plan_visibilities",
		resources.ServicePlanVisibilityResource{},
		func(resource interface{}) bool {
			if spv, ok := resource.(resources.ServicePlanVisibilityResource); ok {
				plans = append(plans, spv.ToFields())
			}
			return true
		})
	return
}

func (repo CloudControllerServicePlanVisibilityRepository) Delete(servicePlanGuid string) error {
	path := fmt.Sprintf("%s/v2/service_plan_visibilities/%s", repo.config.ApiEndpoint(), servicePlanGuid)
	return repo.gateway.DeleteResource(path)
}
