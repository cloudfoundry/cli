package api

import (
	"fmt"
	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
	"net/url"
	"strings"
)

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
		fmt.Sprintf("/v2/organizations/%s/spaces", repo.config.OrganizationFields().Guid),
		resources.SpaceResource{},
		func(resource interface{}) bool {
			return callback(resource.(resources.SpaceResource).ToModel())
		})
}

func (repo CloudControllerSpaceRepository) FindByName(name string) (space models.Space, apiErr error) {
	return repo.FindByNameInOrg(name, repo.config.OrganizationFields().Guid)
}

func (repo CloudControllerSpaceRepository) FindByNameInOrg(name, orgGuid string) (space models.Space, apiErr error) {
	foundSpace := false
	apiErr = repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		fmt.Sprintf("/v2/organizations/%s/spaces?q=%s&inline-relations-depth=1", orgGuid, url.QueryEscape("name:"+strings.ToLower(name))),
		resources.SpaceResource{},
		func(resource interface{}) bool {
			space = resource.(resources.SpaceResource).ToModel()
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
	resource := new(resources.SpaceResource)
	apiErr = repo.gateway.CreateResource(path, strings.NewReader(body), resource)
	if apiErr != nil {
		return
	}
	space = resource.ToModel()
	return
}

func (repo CloudControllerSpaceRepository) Rename(spaceGuid, newName string) (apiErr error) {
	path := fmt.Sprintf("%s/v2/spaces/%s", repo.config.ApiEndpoint(), spaceGuid)
	body := fmt.Sprintf(`{"name":"%s"}`, newName)
	return repo.gateway.UpdateResource(path, strings.NewReader(body))
}

func (repo CloudControllerSpaceRepository) Delete(spaceGuid string) (apiErr error) {
	path := fmt.Sprintf("%s/v2/spaces/%s?recursive=true", repo.config.ApiEndpoint(), spaceGuid)
	return repo.gateway.DeleteResource(path)
}
