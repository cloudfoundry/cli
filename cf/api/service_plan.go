package api

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
)

type ServicePlanRepository interface {
	Search(searchParameters map[string]string) ([]models.ServicePlanFields, error)
}

type CloudControllerServicePlanRepository struct {
	config  configuration.Reader
	gateway net.Gateway
}

func NewCloudControllerServicePlanRepository(config configuration.Reader, gateway net.Gateway) CloudControllerServicePlanRepository {
	return CloudControllerServicePlanRepository{
		config:  config,
		gateway: gateway,
	}
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
		params = append(params, fmt.Sprintf("q=%s", url.QueryEscape(key+":"+value)))
	}

	return uri + "?" + strings.Join(params, "&")
}
