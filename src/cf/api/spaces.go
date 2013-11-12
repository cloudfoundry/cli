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
	NextUrl   string `json:"next_url"`
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
	ListSpaces(stop chan bool) (spacesChan chan []cf.Space, statusChan chan net.ApiResponse)
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

func (repo CloudControllerSpaceRepository) ListSpaces(stop chan bool) (spacesChan chan []cf.Space, statusChan chan net.ApiResponse) {
	spacesChan = make(chan []cf.Space, 4)
	statusChan = make(chan net.ApiResponse, 1)

	go func() {
		path := fmt.Sprintf("/v2/organizations/%s/spaces", repo.config.Organization.Guid)

	loop:
		for path != "" {
			select {
			case <-stop:
				break loop
			default:
				var (
					spaces      []cf.Space
					apiResponse net.ApiResponse
				)
				spaces, path, apiResponse = repo.findNextWithPath(path)
				if apiResponse.IsNotSuccessful() {
					statusChan <- apiResponse
					close(spacesChan)
					close(statusChan)
					return
				}

				spacesChan <- spaces
			}
		}
		close(spacesChan)
		close(statusChan)
		cf.WaitForClose(stop)
	}()

	return
}

func (repo CloudControllerSpaceRepository) FindByName(name string) (space cf.Space, apiResponse net.ApiResponse) {
	return repo.FindByNameInOrg(name, repo.config.Organization)
}

func (repo CloudControllerSpaceRepository) FindByNameInOrg(name string, org cf.Organization) (space cf.Space, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("/v2/organizations/%s/spaces?q=name%%3A%s&inline-relations-depth=1", org.Guid, strings.ToLower(name))

	spaces, _, apiResponse := repo.findNextWithPath(path)
	if apiResponse.IsNotSuccessful() {
		return
	}

	if len(spaces) == 0 {
		apiResponse = net.NewNotFoundApiResponse("%s %s not found", "Space", name)
		return
	}

	space = spaces[0]
	return
}

func (repo CloudControllerSpaceRepository) findNextWithPath(path string) (spaces []cf.Space, nextUrl string, apiResponse net.ApiResponse) {
	resources := new(PaginatedSpaceResources)
	apiResponse = repo.gateway.GetResource(repo.config.Target+path, repo.config.AccessToken, resources)
	if apiResponse.IsNotSuccessful() {
		return
	}

	nextUrl = resources.NextUrl

	for _, r := range resources.Resources {
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
		space := cf.Space{
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

		spaces = append(spaces, space)
	}
	return
}

func (repo CloudControllerSpaceRepository) Create(name string) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/spaces", repo.config.Target)
	body := fmt.Sprintf(`{"name":"%s","organization_guid":"%s"}`, name, repo.config.Organization.Guid)
	return repo.gateway.CreateResource(path, repo.config.AccessToken, strings.NewReader(body))
}

func (repo CloudControllerSpaceRepository) Rename(space cf.Space, newName string) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/spaces/%s", repo.config.Target, space.Guid)
	body := fmt.Sprintf(`{"name":"%s"}`, newName)
	return repo.gateway.UpdateResource(path, repo.config.AccessToken, strings.NewReader(body))
}

func (repo CloudControllerSpaceRepository) Delete(space cf.Space) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/spaces/%s?recursive=true", repo.config.Target, space.Guid)
	return repo.gateway.DeleteResource(path, repo.config.AccessToken)
}
