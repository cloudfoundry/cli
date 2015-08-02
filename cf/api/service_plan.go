package api

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
)

type ServicePlanRepository interface {
	Search(searchParameters map[string]string) ([]models.ServicePlanFields, error)
	Update(models.ServicePlanFields, string, bool) error
	ListPlansFromManyServices(serviceGuids []string) ([]models.ServicePlanFields, error)
}

type CloudControllerServicePlanRepository struct {
	config  core_config.Reader
	gateway net.Gateway
}

func NewCloudControllerServicePlanRepository(config core_config.Reader, gateway net.Gateway) CloudControllerServicePlanRepository {
	return CloudControllerServicePlanRepository{
		config:  config,
		gateway: gateway,
	}
}

func (repo CloudControllerServicePlanRepository) Update(servicePlan models.ServicePlanFields, serviceGuid string, public bool) error {
	var body string

	body = fmt.Sprintf(`{"name":"%s", "free":%t, "description":"%s", "public":%t, "service_guid":"%s"}`,
		servicePlan.Name,
		servicePlan.Free,
		servicePlan.Description,
		public,
		serviceGuid,
	)

	url := fmt.Sprintf("/v2/service_plans/%s", servicePlan.Guid)
	return repo.gateway.UpdateResource(repo.config.ApiEndpoint(), url, strings.NewReader(body))
}

func (repo CloudControllerServicePlanRepository) ListPlansFromManyServices(serviceGuids []string) ([]models.ServicePlanFields, error) {
	serviceGuidsString := strings.Join(serviceGuids, ",")
	plans := []models.ServicePlanFields{}

	err := repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		fmt.Sprintf("/v2/service_plans?q=%s", url.QueryEscape("service_guid IN "+serviceGuidsString)),
		resources.ServicePlanResource{},
		func(resource interface{}) bool {
			if plan, ok := resource.(resources.ServicePlanResource); ok {
				plans = append(plans, plan.ToFields())
			}
			return true
		})
	return plans, err
}

func (repo CloudControllerServicePlanRepository) Search(queryParams map[string]string) (plans []models.ServicePlanFields, err error) {
	err = repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		combineQueryParametersWithUri("/v2/service_plans", queryParams),
		resources.ServicePlanResource{},
		func(resource interface{}) bool {
			if sp, ok := resource.(resources.ServicePlanResource); ok {
				plans = append(plans, sp.ToFields())
			}
			return true
		})
	return
}

func combineQueryParametersWithUri(uri string, queryParams map[string]string) string {
	if len(queryParams) == 0 {
		return uri
	}

	params := []string{}
	for key, value := range queryParams {
		params = append(params, url.QueryEscape(key+":"+value))
	}

	return uri + "?q=" + strings.Join(params, "%3B")
}
