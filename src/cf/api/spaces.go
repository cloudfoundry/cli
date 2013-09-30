package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"strings"
)

type SpaceRepository interface {
	GetCurrentSpace() (space cf.Space)
	FindAll() (spaces []cf.Space, apiErr *net.ApiError)
	FindByName(name string) (space cf.Space, apiErr *net.ApiError)
	GetSummary() (space cf.Space, apiErr *net.ApiError)
	Create(name string) (apiErr *net.ApiError)
	Rename(space cf.Space, newName string) (apiErr *net.ApiError)
	Delete(space cf.Space) (apiErr *net.ApiError)
}

type CloudControllerSpaceRepository struct {
	config  configuration.Configuration
	gateway net.Gateway
}

func NewCloudControllerSpaceRepository(config configuration.Configuration, gateway net.Gateway) (repo CloudControllerSpaceRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerSpaceRepository) GetCurrentSpace() (space cf.Space) {
	return repo.config.Space
}

func (repo CloudControllerSpaceRepository) FindAll() (spaces []cf.Space, apiErr *net.ApiError) {
	path := fmt.Sprintf("%s/v2/organizations/%s/spaces", repo.config.Target, repo.config.Organization.Guid)
	request, apiErr := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiErr != nil {
		return
	}

	response := new(ApiResponse)

	_, apiErr = repo.gateway.PerformRequestForJSONResponse(request, response)

	if apiErr != nil {
		return
	}

	for _, r := range response.Resources {
		spaces = append(spaces, cf.Space{Name: r.Entity.Name, Guid: r.Metadata.Guid})
	}

	return
}

func (repo CloudControllerSpaceRepository) FindByName(name string) (space cf.Space, apiErr *net.ApiError) {
	path := fmt.Sprintf("%s/v2/spaces?q=name%s&inline-relations-depth=1", repo.config.Target, "%3A"+strings.ToLower(name))

	request, apiErr := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiErr != nil {
		return
	}

	response := new(SpaceApiResponse)

	_, apiErr = repo.gateway.PerformRequestForJSONResponse(request, response)

	if apiErr != nil {
		return
	}

	if len(response.Resources) == 0 {
		return
	}

	r := response.Resources[0]
	apps := []cf.Application{}
	for _, app := range r.Entity.Applications {
		apps = append(apps, cf.Application{Name: app.Entity.Name, Guid: app.Metadata.Guid})
	}

	domains := []cf.Domain{}
	for _, domain := range r.Entity.Domains {
		domains = append(domains, cf.Domain{Name: domain.Entity.Name, Guid: domain.Metadata.Guid})
	}

	services := []cf.ServiceInstance{}
	for _, service := range r.Entity.ServiceInstances {
		services = append(services, cf.ServiceInstance{Name: service.Entity.Name, Guid: service.Metadata.Guid})
	}
	space = cf.Space{
		Name: r.Entity.Name,
		Guid: r.Metadata.Guid,
		Organization: cf.Organization{
			Name: r.Entity.Organization.Entity.Name,
			Guid: r.Entity.Organization.Metadata.Guid,
		},
		Applications:     apps,
		Domains:          domains,
		ServiceInstances: services,
	}
	return
}

func (repo CloudControllerSpaceRepository) GetSummary() (space cf.Space, apiErr *net.ApiError) {
	path := fmt.Sprintf("%s/v2/spaces/%s/summary", repo.config.Target, repo.config.Space.Guid)
	request, apiErr := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiErr != nil {
		return
	}

	response := new(SpaceSummary) // but not an ApiResponse
	_, apiErr = repo.gateway.PerformRequestForJSONResponse(request, response)

	if apiErr != nil {
		return
	}

	applications := extractApplicationsFromSummary(response.Apps)
	serviceInstances := extractServiceInstancesFromSummary(response.ServiceInstances, response.Apps)

	space = cf.Space{Name: response.Name, Guid: response.Guid, Applications: applications, ServiceInstances: serviceInstances}

	return
}

func (repo CloudControllerSpaceRepository) Create(name string) (apiErr *net.ApiError) {
	path := fmt.Sprintf("%s/v2/spaces", repo.config.Target)
	body := fmt.Sprintf(`{"name":"%s","organization_guid":"%s"}`, name, repo.config.Organization.Guid)

	request, apiErr := repo.gateway.NewRequest("POST", path, repo.config.AccessToken, strings.NewReader(body))
	if apiErr != nil {
		return
	}

	apiErr = repo.gateway.PerformRequest(request)
	return
}

func (repo CloudControllerSpaceRepository) Rename(space cf.Space, newName string) (apiErr *net.ApiError) {
	path := fmt.Sprintf("%s/v2/spaces/%s", repo.config.Target, space.Guid)
	body := fmt.Sprintf(`{"name":"%s"}`, newName)

	request, apiErr := repo.gateway.NewRequest("PUT", path, repo.config.AccessToken, strings.NewReader(body))
	if apiErr != nil {
		return
	}

	apiErr = repo.gateway.PerformRequest(request)
	return
}

func (repo CloudControllerSpaceRepository) Delete(space cf.Space) (apiErr *net.ApiError) {
	path := fmt.Sprintf("%s/v2/spaces/%s?recursive=true", repo.config.Target, space.Guid)

	request, apiErr := repo.gateway.NewRequest("DELETE", path, repo.config.AccessToken, nil)
	if apiErr != nil {
		return
	}

	apiErr = repo.gateway.PerformRequest(request)
	return
}

func extractApplicationsFromSummary(appSummaries []ApplicationSummary) (applications []cf.Application) {
	for _, appSummary := range appSummaries {
		app := cf.Application{
			Name:             appSummary.Name,
			Guid:             appSummary.Guid,
			Urls:             appSummary.Urls,
			State:            strings.ToLower(appSummary.State),
			Instances:        appSummary.Instances,
			RunningInstances: appSummary.RunningInstances,
			Memory:           appSummary.Memory,
		}
		applications = append(applications, app)
	}

	return
}

func extractServiceInstancesFromSummary(instanceSummaries []ServiceInstanceSummary, appSummaries []ApplicationSummary) (instances []cf.ServiceInstance) {
	for _, instanceSummary := range instanceSummaries {
		applicationNames := findApplicationNamesForInstance(instanceSummary.Name, appSummaries)

		planSummary := instanceSummary.ServicePlan
		offeringSummary := planSummary.ServiceOffering

		serviceOffering := cf.ServiceOffering{
			Label:    offeringSummary.Label,
			Provider: offeringSummary.Provider,
			Version:  offeringSummary.Version,
		}

		servicePlan := cf.ServicePlan{
			Name:            planSummary.Name,
			ServiceOffering: serviceOffering,
		}

		instance := cf.ServiceInstance{
			Name:             instanceSummary.Name,
			ServicePlan:      servicePlan,
			ApplicationNames: applicationNames,
		}

		instances = append(instances, instance)
	}

	return
}

func findApplicationNamesForInstance(instanceName string, appSummaries []ApplicationSummary) (applicationNames []string) {
	for _, appSummary := range appSummaries {
		for _, name := range appSummary.ServiceNames {
			if name == instanceName {
				applicationNames = append(applicationNames, appSummary.Name)
			}
		}
	}

	return
}
