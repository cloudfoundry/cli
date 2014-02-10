package api

import (
	"cf/configuration"
	"cf/models"
	"cf/net"
	"fmt"
	"net/url"
	"strings"
)

type SpaceResource struct {
	Metadata Metadata
	Entity   SpaceEntity
}

func (resource SpaceResource) ToFields() (fields models.SpaceFields) {
	fields.Guid = resource.Metadata.Guid
	fields.Name = resource.Entity.Name
	return
}

func (resource SpaceResource) ToModel() (space models.Space) {
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
	ListSpaces(func(models.Space) bool) net.ApiResponse
	FindByName(name string) (space models.Space, apiResponse net.ApiResponse)
	FindByNameInOrg(name, orgGuid string) (space models.Space, apiResponse net.ApiResponse)
	Create(name string, orgGuid string) (space models.Space, apiResponse net.ApiResponse)
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

func (repo CloudControllerSpaceRepository) ListSpaces(callback func(models.Space) bool) net.ApiResponse {
	return repo.gateway.ListPaginatedResources(
		repo.config.Target,
		repo.config.AccessToken,
		fmt.Sprintf("/v2/organizations/%s/spaces", repo.config.OrganizationFields.Guid),
		SpaceResource{},
		func(resource interface{}) bool {
			return callback(resource.(SpaceResource).ToModel())
		})
}

func (repo CloudControllerSpaceRepository) FindByName(name string) (space models.Space, apiResponse net.ApiResponse) {
	return repo.FindByNameInOrg(name, repo.config.OrganizationFields.Guid)
}

func (repo CloudControllerSpaceRepository) FindByNameInOrg(name, orgGuid string) (space models.Space, apiResponse net.ApiResponse) {
	foundSpace := false
	apiResponse = repo.gateway.ListPaginatedResources(
		repo.config.Target,
		repo.config.AccessToken,
		fmt.Sprintf("/v2/organizations/%s/spaces?q=%s&inline-relations-depth=1", orgGuid, url.QueryEscape("name:"+strings.ToLower(name))),
		SpaceResource{},
		func(resource interface{}) bool {
			space = resource.(SpaceResource).ToModel()
			foundSpace = true
			return false
		})

	if !foundSpace {
		apiResponse = net.NewNotFoundApiResponse("Space %s not found.", name)
	}

	return
}

func (repo CloudControllerSpaceRepository) Create(name string, orgGuid string) (space models.Space, apiResponse net.ApiResponse) {
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
