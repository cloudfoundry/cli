package spaces

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
)

type SpaceRepository interface {
	ListSpaces(func(models.Space) bool) error
	FindByName(name string) (space models.Space, apiErr error)
	FindByNameInOrg(name, orgGuid string) (space models.Space, apiErr error)
	Create(name string, orgGuid string, spaceQuotaGuid string) (space models.Space, apiErr error)
	Rename(spaceGuid, newName string) (apiErr error)
	Delete(spaceGuid string) (apiErr error)
}

type CloudControllerSpaceRepository struct {
	config  core_config.Reader
	gateway net.Gateway
}

func NewCloudControllerSpaceRepository(config core_config.Reader, gateway net.Gateway) (repo CloudControllerSpaceRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerSpaceRepository) ListSpaces(callback func(models.Space) bool) error {
	return repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		fmt.Sprintf("/v2/organizations/%s/spaces?inline-relations-depth=1", repo.config.OrganizationFields().Guid),
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

func (repo CloudControllerSpaceRepository) Create(name, orgGuid, spaceQuotaGuid string) (space models.Space, apiErr error) {
	path := "/v2/spaces?inline-relations-depth=1"

	bodyMap := map[string]string{"name": name, "organization_guid": orgGuid}
	if spaceQuotaGuid != "" {
		bodyMap["space_quota_definition_guid"] = spaceQuotaGuid
	}

	body, apiErr := json.Marshal(bodyMap)
	if apiErr != nil {
		return
	}

	resource := new(resources.SpaceResource)
	apiErr = repo.gateway.CreateResource(repo.config.ApiEndpoint(), path, strings.NewReader(string(body)), resource)
	if apiErr != nil {
		return
	}
	space = resource.ToModel()
	return
}

func (repo CloudControllerSpaceRepository) Rename(spaceGuid, newName string) (apiErr error) {
	path := fmt.Sprintf("/v2/spaces/%s", spaceGuid)
	body := fmt.Sprintf(`{"name":"%s"}`, newName)
	return repo.gateway.UpdateResource(repo.config.ApiEndpoint(), path, strings.NewReader(body))
}

func (repo CloudControllerSpaceRepository) Delete(spaceGuid string) (apiErr error) {
	path := fmt.Sprintf("/v2/spaces/%s?recursive=true", spaceGuid)
	return repo.gateway.DeleteResource(repo.config.ApiEndpoint(), path)
}
