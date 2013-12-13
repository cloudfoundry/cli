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

func (resource SpaceResource) ToFields() (fields cf.SpaceFields) {
	fields.Guid = resource.Metadata.Guid
	fields.Name = resource.Entity.Name
	return
}

func (resource SpaceResource) ToModel() (space cf.Space) {
	space.SpaceFields = resource.ToFields()
	for _, app := range resource.Entity.Applications {
		space.Applications = append(space.Applications, app.ToFields())
	}

	for _, domainResource := range resource.Entity.Domains {
		space.Domains = append(space.Domains, domainResource.ToFields())
	}

	for _, serviceResource := range resource.Entity.ServiceInstances {
		space.ServiceInstances = append(space.ServiceInstances, serviceResource.ToFields())
	}

	space.Organization = resource.Entity.Organization.ToFields()
	return
}

type SpaceEntity struct {
	Name             string
	Organization     OrganizationResource
	Applications     []ApplicationResource `json:"apps"`
	Domains          []DomainResource
	ServiceInstances []ServiceInstanceResource `json:"service_instances"`
}

type SpaceRepository interface {
	ListSpaces(stop chan bool) (spacesChan chan []cf.Space, statusChan chan net.ApiResponse)
	FindByName(name string) (space cf.Space, apiResponse net.ApiResponse)
	FindByNameInOrg(name, orgGuid string) (space cf.Space, apiResponse net.ApiResponse)
	Create(name string, orgGuid string) (space cf.Space, apiResponse net.ApiResponse)
	Rename(spaceGuid, newName string) (apiResponse net.ApiResponse)
	Delete(spaceGuid string) (apiResponse net.ApiResponse)
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
		path := fmt.Sprintf("/v2/organizations/%s/spaces", repo.config.OrganizationFields.Guid)

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

				if len(spaces) > 0 {
					spacesChan <- spaces
				}
			}
		}
		close(spacesChan)
		close(statusChan)
		cf.WaitForClose(stop)
	}()

	return
}

func (repo CloudControllerSpaceRepository) FindByName(name string) (space cf.Space, apiResponse net.ApiResponse) {
	return repo.FindByNameInOrg(name, repo.config.OrganizationFields.Guid)
}

func (repo CloudControllerSpaceRepository) FindByNameInOrg(name, orgGuid string) (space cf.Space, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("/v2/organizations/%s/spaces?q=name%%3A%s&inline-relations-depth=1", orgGuid, strings.ToLower(name))

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
		spaces = append(spaces, r.ToModel())
	}
	return
}

func (repo CloudControllerSpaceRepository) Create(name string, orgGuid string) (space cf.Space, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/spaces?inline-relations-depth=1", repo.config.Target)
	body := fmt.Sprintf(`{"name":"%s","organization_guid":"%s"}`, name, orgGuid)
	resource := new(SpaceResource)
	apiResponse = repo.gateway.CreateResourceForResponse(path, repo.config.AccessToken, strings.NewReader(body), resource)
	if apiResponse.IsNotSuccessful() {
		return
	}
	space = resource.ToModel()
	return
}

func (repo CloudControllerSpaceRepository) Rename(spaceGuid, newName string) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/spaces/%s", repo.config.Target, spaceGuid)
	body := fmt.Sprintf(`{"name":"%s"}`, newName)
	return repo.gateway.UpdateResource(path, repo.config.AccessToken, strings.NewReader(body))
}

func (repo CloudControllerSpaceRepository) Delete(spaceGuid string) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/spaces/%s?recursive=true", repo.config.Target, spaceGuid)
	return repo.gateway.DeleteResource(path, repo.config.AccessToken)
}
