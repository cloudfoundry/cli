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
	Search(map[string]string) ([]models.ServicePlanVisibilityFields, error)
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
	path := fmt.Sprintf("%s/v2/service_plan_visibilities/%s", repo.config.ApiEndpoint(), servicePlanGuid)
	return repo.gateway.DeleteResource(path)
}

func (repo CloudControllerServicePlanVisibilityRepository) Search(queryParams map[string]string) ([]models.ServicePlanVisibilityFields, error) {
	var visibilities []models.ServicePlanVisibilityFields
	var filteredVisibilities []models.ServicePlanVisibilityFields
	err := repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		"/v2/service_plan_visibilities", // This will eventually mix in query parameters
		resources.ServicePlanVisibilityResource{},
		func(resource interface{}) bool {
			if sp, ok := resource.(resources.ServicePlanVisibilityResource); ok {
				visibilities = append(visibilities, sp.ToFields())
			}
			return true
		})

	//WE DO THE HORRIBLE BROKEN NAIVE THING HERE BECAUSE THE SERVICES API DOES NOT SUPPORT QUERY PARAMS
	//AS OF 8/11/2014.  AS SOON AS IT DOES, THIS HACKY FILTERING SHOULD BE REMOVED.
	for _, visibility := range visibilities {
		if visibility.OrganizationGuid == queryParams["org_guid"] && visibility.ServicePlanGuid == queryParams["plan_guid"] {
			filteredVisibilities = append(filteredVisibilities, visibility)
		}
	}
	return filteredVisibilities, err

}
