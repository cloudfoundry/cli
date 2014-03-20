package api

import (
	"cf/configuration"
	"cf/errors"
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
	ListSpaces(func(models.Space) bool) error
	FindByName(name string) (space models.Space, apiErr error)
	FindByNameInOrg(name, orgGuid string) (space models.Space, apiErr error)
	Create(name string, orgGuid string) (space models.Space, apiErr error)
	Rename(spaceGuid, newName string) (apiErr error)
	Delete(spaceGuid string) (apiErr error)
}

type CloudControllerSpaceRepository struct {
	config  configuration.Reader
	gateway net.Gateway
}

func NewCloudControllerSpaceRepository(config configuration.Reader, gateway net.Gateway) (repo CloudControllerSpaceRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerSpaceRepository) ListSpaces(callback func(models.Space) bool) error {
	return repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		repo.config.AccessToken(),
		fmt.Sprintf("/v2/organizations/%s/spaces", repo.config.OrganizationFields().Guid),
		SpaceResource{},
		func(resource interface{}) bool {
			return callback(resource.(SpaceResource).ToModel())
		})
}

func (repo CloudControllerSpaceRepository) FindByName(name string) (space models.Space, apiErr error) {
	return repo.FindByNameInOrg(name, repo.config.OrganizationFields().Guid)
}

func (repo CloudControllerSpaceRepository) FindByNameInOrg(name, orgGuid string) (space models.Space, apiErr error) {
	foundSpace := false
	apiErr = repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		repo.config.AccessToken(),
		fmt.Sprintf("/v2/organizations/%s/spaces?q=%s&inline-relations-depth=1", orgGuid, url.QueryEscape("name:"+strings.ToLower(name))),
		SpaceResource{},
		func(resource interface{}) bool {
			space = resource.(SpaceResource).ToModel()
			foundSpace = true
			return false
		})

	if !foundSpace {
		apiErr = errors.NewModelNotFoundError("Space", name)
	}

	return
}

func (repo CloudControllerSpaceRepository) Create(name string, orgGuid string) (space models.Space, apiErr error) {
	path := fmt.Sprintf("%s/v2/spaces?inline-relations-depth=1", repo.config.ApiEndpoint())
	body := fmt.Sprintf(`{"name":"%s","organization_guid":"%s"}`, name, orgGuid)
	resource := new(SpaceResource)
	apiErr = repo.gateway.CreateResourceForResponse(path, repo.config.AccessToken(), strings.NewReader(body), resource)
	if apiErr != nil {
		return
	}
	space = resource.ToModel()
	return
}

func (repo CloudControllerSpaceRepository) Rename(spaceGuid, newName string) (apiErr error) {
	path := fmt.Sprintf("%s/v2/spaces/%s", repo.config.ApiEndpoint(), spaceGuid)
	body := fmt.Sprintf(`{"name":"%s"}`, newName)
	return repo.gateway.UpdateResource(path, repo.config.AccessToken(), strings.NewReader(body))
}

func (repo CloudControllerSpaceRepository) Delete(spaceGuid string) (apiErr error) {
	path := fmt.Sprintf("%s/v2/spaces/%s?recursive=true", repo.config.ApiEndpoint(), spaceGuid)
	return repo.gateway.DeleteResource(path, repo.config.AccessToken())
}
