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

type OrganizationRepository interface {
	ListOrgs(func(models.Organization) bool) (apiErr error)
	FindByName(name string) (org models.Organization, apiErr error)
	Create(name string) (apiErr error)
	Rename(orgGuid string, name string) (apiErr error)
	Delete(orgGuid string) (apiErr error)
}

type CloudControllerOrganizationRepository struct {
	config  configuration.Reader
	gateway net.Gateway
}

func NewCloudControllerOrganizationRepository(config configuration.Reader, gateway net.Gateway) (repo CloudControllerOrganizationRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerOrganizationRepository) ListOrgs(cb func(models.Organization) bool) (apiErr error) {
	return repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		"/v2/organizations",
		resources.OrganizationResource{},
		func(resource interface{}) bool {
			return cb(resource.(resources.OrganizationResource).ToModel())
		})
}

func (repo CloudControllerOrganizationRepository) FindByName(name string) (org models.Organization, apiErr error) {
	found := false
	apiErr = repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		fmt.Sprintf("/v2/organizations?q=%s&inline-relations-depth=1", url.QueryEscape("name:"+strings.ToLower(name))),
		resources.OrganizationResource{},
		func(resource interface{}) bool {
			org = resource.(resources.OrganizationResource).ToModel()
			found = true
			return false
		})

	if apiErr == nil && !found {
		apiErr = errors.NewModelNotFoundError("Organization", name)
	}

	return
}

func (repo CloudControllerOrganizationRepository) Create(name string) (apiErr error) {
	url := repo.config.ApiEndpoint() + "/v2/organizations"
	data := fmt.Sprintf(`{"name":"%s"}`, name)
	return repo.gateway.CreateResource(url, strings.NewReader(data))
}

func (repo CloudControllerOrganizationRepository) Rename(orgGuid string, name string) (apiErr error) {
	url := fmt.Sprintf("%s/v2/organizations/%s", repo.config.ApiEndpoint(), orgGuid)
	data := fmt.Sprintf(`{"name":"%s"}`, name)
	return repo.gateway.UpdateResource(url, strings.NewReader(data))
}

func (repo CloudControllerOrganizationRepository) Delete(orgGuid string) (apiErr error) {
	url := fmt.Sprintf("%s/v2/organizations/%s?recursive=true", repo.config.ApiEndpoint(), orgGuid)
	return repo.gateway.DeleteResource(url)
}
