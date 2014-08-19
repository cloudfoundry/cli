package space_quotas

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
)

type SpaceQuotaRepository interface {
	FindAll() (quotas []models.SpaceQuota, apiErr error)
	FindByName(name string) (quota models.SpaceQuota, apiErr error)
	FindByOrg(guid string) (quota []models.SpaceQuota, apiErr error)

	AssignQuotaToOrg(orgGuid, quotaGuid string) error

	// CRUD ahoy
	Create(quota models.SpaceQuota) error
	Update(quota models.SpaceQuota) error
	Delete(quotaGuid string) error
}

type CloudControllerSpaceQuotaRepository struct {
	config  configuration.Reader
	gateway net.Gateway
}

func NewCloudControllerSpaceQuotaRepository(config configuration.Reader, gateway net.Gateway) (repo CloudControllerSpaceQuotaRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerSpaceQuotaRepository) findAllWithPath(path string) ([]models.SpaceQuota, error) {
	var quotas []models.SpaceQuota
	apiErr := repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		path,
		resources.SpaceQuotaResource{},
		func(resource interface{}) bool {
			if qr, ok := resource.(resources.SpaceQuotaResource); ok {
				quotas = append(quotas, qr.ToModel())
			}
			return true
		})
	return quotas, apiErr
}

func (repo CloudControllerSpaceQuotaRepository) FindAll() (quotas []models.SpaceQuota, apiErr error) {
	return repo.findAllWithPath("/v2/quota_definitions")
}

func (repo CloudControllerSpaceQuotaRepository) FindByName(name string) (quota models.SpaceQuota, apiErr error) {
	quotas, apiErr := repo.FindByOrg(repo.config.OrganizationFields().Guid)
	if apiErr != nil {
		return
	}

	for _, quota := range quotas {
		if quota.Name == name {
			return quota, nil
		}
	}

	apiErr = errors.NewModelNotFoundError("Space Quota", name)
	return models.SpaceQuota{}, apiErr
}

func (repo CloudControllerSpaceQuotaRepository) FindByOrg(guid string) ([]models.SpaceQuota, error) {
	path := fmt.Sprintf("/v2/organizations/%s/space_quota_definitions", guid)
	quotas, apiErr := repo.findAllWithPath(path)
	if apiErr != nil {
		return nil, apiErr
	}
	return quotas, nil
}

func (repo CloudControllerSpaceQuotaRepository) Create(quota models.SpaceQuota) error {
	path := fmt.Sprintf("%s/v2/space_quota_definitions", repo.config.ApiEndpoint())
	return repo.gateway.CreateResourceFromStruct(path, quota)
}

func (repo CloudControllerSpaceQuotaRepository) Update(quota models.SpaceQuota) error {
	path := fmt.Sprintf("%s/v2/quota_definitions/%s", repo.config.ApiEndpoint(), quota.Guid)
	return repo.gateway.UpdateResourceFromStruct(path, quota)
}

func (repo CloudControllerSpaceQuotaRepository) AssignQuotaToOrg(orgGuid, quotaGuid string) (apiErr error) {
	path := fmt.Sprintf("%s/v2/organizations/%s", repo.config.ApiEndpoint(), orgGuid)
	data := fmt.Sprintf(`{"quota_definition_guid":"%s"}`, quotaGuid)
	return repo.gateway.UpdateResource(path, strings.NewReader(data))
}

func (repo CloudControllerSpaceQuotaRepository) Delete(quotaGuid string) (apiErr error) {
	path := fmt.Sprintf("%s/v2/quota_definitions/%s", repo.config.ApiEndpoint(), quotaGuid)
	return repo.gateway.DeleteResource(path)
}
