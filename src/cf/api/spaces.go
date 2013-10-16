package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"strings"
)

type PaginatedSpaceResources struct {
	Resources []SpaceResource
}

type SpaceResource struct {
	Metadata Metadata
	Entity   SpaceEntity
}

type SpaceEntity struct {
	Name             string
	Organization     Resource
	Applications     []Resource `json:"apps"`
	Domains          []Resource
	ServiceInstances []Resource `json:"service_instances"`
}

type SpaceRepository interface {
	FindAll() (spaces []cf.Space, apiResponse net.ApiResponse)
	FindByName(name string) (space cf.Space, apiResponse net.ApiResponse)
	FindByNameInOrg(name string, org cf.Organization) (space cf.Space, apiResponse net.ApiResponse)
	Create(name string) (apiResponse net.ApiResponse)
	Rename(space cf.Space, newName string) (apiResponse net.ApiResponse)
	Delete(space cf.Space) (apiResponse net.ApiResponse)
}

type CloudControllerSpaceRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
}

func NewCloudControllerSpaceRepository(config *configuration.Configuration, gateway net.Gateway) (repo CloudControllerSpaceRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerSpaceRepository) FindAll() (spaces []cf.Space, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/organizations/%s/spaces", repo.config.Target, repo.config.Organization.Guid)
	request, apiResponse := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	resources := new(PaginatedResources)

	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, resources)

	if apiResponse.IsNotSuccessful() {
		return
	}

	for _, r := range resources.Resources {
		spaces = append(spaces, cf.Space{Name: r.Entity.Name, Guid: r.Metadata.Guid})
	}

	return
}

func (repo CloudControllerSpaceRepository) FindByName(name string) (space cf.Space, apiResponse net.ApiResponse) {
	return repo.FindByNameInOrg(name, repo.config.Organization)
}

func (repo CloudControllerSpaceRepository) FindByNameInOrg(name string, org cf.Organization) (space cf.Space, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/organizations/%s/spaces?q=name%s&inline-relations-depth=1",
		repo.config.Target, org.Guid, "%3A"+strings.ToLower(name))

	request, apiResponse := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	resources := new(PaginatedSpaceResources)

	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, resources)

	if apiResponse.IsNotSuccessful() {
		return
	}

	if len(resources.Resources) == 0 {
		apiResponse = net.NewNotFoundApiResponse("%s %s not found", "Space", name)
		return
	}

	r := resources.Resources[0]
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

func (repo CloudControllerSpaceRepository) Create(name string) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/spaces", repo.config.Target)
	body := fmt.Sprintf(`{"name":"%s","organization_guid":"%s"}`, name, repo.config.Organization.Guid)

	request, apiResponse := repo.gateway.NewRequest("POST", path, repo.config.AccessToken, strings.NewReader(body))
	if apiResponse.IsNotSuccessful() {
		return
	}

	apiResponse = repo.gateway.PerformRequest(request)
	return
}

func (repo CloudControllerSpaceRepository) Rename(space cf.Space, newName string) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/spaces/%s", repo.config.Target, space.Guid)
	body := fmt.Sprintf(`{"name":"%s"}`, newName)

	request, apiResponse := repo.gateway.NewRequest("PUT", path, repo.config.AccessToken, strings.NewReader(body))
	if apiResponse.IsNotSuccessful() {
		return
	}

	apiResponse = repo.gateway.PerformRequest(request)
	return
}

func (repo CloudControllerSpaceRepository) Delete(space cf.Space) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/spaces/%s?recursive=true", repo.config.Target, space.Guid)

	request, apiResponse := repo.gateway.NewRequest("DELETE", path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	apiResponse = repo.gateway.PerformRequest(request)
	return
}
