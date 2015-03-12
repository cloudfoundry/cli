package organizations

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
)

//go:generate counterfeiter -o fakes/fake_organization_repository.go . OrganizationRepository
type OrganizationRepository interface {
	ListOrgs() (orgs []models.Organization, apiErr error)
	FindByName(name string) (org models.Organization, apiErr error)
	Create(org models.Organization) (apiErr error)
	Rename(orgGuid string, name string) (apiErr error)
	Delete(orgGuid string) (apiErr error)
	SharePrivateDomain(orgGuid string, domainGuid string) (apiErr error)
	UnsharePrivateDomain(orgGuid string, domainGuid string) (apiErr error)
}

type CloudControllerOrganizationRepository struct {
	config  core_config.Reader
	gateway net.Gateway
}

func NewCloudControllerOrganizationRepository(config core_config.Reader, gateway net.Gateway) (repo CloudControllerOrganizationRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerOrganizationRepository) ListOrgs() ([]models.Organization, error) {
	orgs := []models.Organization{}
	err := repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		"/v2/organizations",
		resources.OrganizationResource{},
		func(resource interface{}) bool {
			orgResource, ok := resource.(resources.OrganizationResource)
			if ok {
				orgs = append(orgs, orgResource.ToModel())
				return true
			} else {
				return false
			}
		})
	return orgs, err
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

func (repo CloudControllerOrganizationRepository) Create(org models.Organization) (apiErr error) {
	data := fmt.Sprintf(`{"name":"%s"`, org.Name)
	if org.QuotaDefinition.Guid != "" {
		data = data + fmt.Sprintf(`, "quota_definition_guid":"%s"`, org.QuotaDefinition.Guid)
	}
	data = data + "}"
	return repo.gateway.CreateResource(repo.config.ApiEndpoint(), "/v2/organizations", strings.NewReader(data))
}

func (repo CloudControllerOrganizationRepository) Rename(orgGuid string, name string) (apiErr error) {
	url := fmt.Sprintf("/v2/organizations/%s", orgGuid)
	data := fmt.Sprintf(`{"name":"%s"}`, name)
	return repo.gateway.UpdateResource(repo.config.ApiEndpoint(), url, strings.NewReader(data))
}

func (repo CloudControllerOrganizationRepository) Delete(orgGuid string) (apiErr error) {
	url := fmt.Sprintf("/v2/organizations/%s?recursive=true", orgGuid)
	return repo.gateway.DeleteResource(repo.config.ApiEndpoint(), url)
}

func (repo CloudControllerOrganizationRepository) SharePrivateDomain(orgGuid string, domainGuid string) error {
	url := fmt.Sprintf("/v2/organizations/%s/private_domains/%s", orgGuid, domainGuid)
	return repo.gateway.UpdateResource(repo.config.ApiEndpoint(), url, nil)
}

func (repo CloudControllerOrganizationRepository) UnsharePrivateDomain(orgGuid string, domainGuid string) error {
	url := fmt.Sprintf("/v2/organizations/%s/private_domains/%s", orgGuid, domainGuid)
	return repo.gateway.DeleteResource(repo.config.ApiEndpoint(), url)
}
