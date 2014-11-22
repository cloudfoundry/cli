package space_quotas

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
)

type SpaceQuotaRepository interface {
	FindByName(name string) (quota models.SpaceQuota, apiErr error)
	FindByOrg(guid string) (quota []models.SpaceQuota, apiErr error)
	FindByGuid(guid string) (quota models.SpaceQuota, apiErr error)

	AssociateSpaceWithQuota(spaceGuid string, quotaGuid string) error
	UnassignQuotaFromSpace(spaceGuid string, quotaGuid string) error

	// CRUD ahoy
	Create(quota models.SpaceQuota) error
	Update(quota models.SpaceQuota) error
	Delete(quotaGuid string) error
}

type CloudControllerSpaceQuotaRepository struct {
	config  core_config.Reader
	gateway net.Gateway
}

func NewCloudControllerSpaceQuotaRepository(config core_config.Reader, gateway net.Gateway) (repo CloudControllerSpaceQuotaRepository) {
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

func (repo CloudControllerSpaceQuotaRepository) FindByGuid(guid string) (quota models.SpaceQuota, apiErr error) {
	quotas, apiErr := repo.FindByOrg(repo.config.OrganizationFields().Guid)
	if apiErr != nil {
		return
	}

	for _, quota := range quotas {
		if quota.Guid == guid {
			return quota, nil
		}
	}

	apiErr = errors.NewModelNotFoundError("Space Quota", guid)
	return models.SpaceQuota{}, apiErr
}

func (repo CloudControllerSpaceQuotaRepository) Create(quota models.SpaceQuota) error {
	path := "/v2/space_quota_definitions"
	return repo.gateway.CreateResourceFromStruct(repo.config.ApiEndpoint(), path, quota)
}

func (repo CloudControllerSpaceQuotaRepository) Update(quota models.SpaceQuota) error {
	path := fmt.Sprintf("/v2/space_quota_definitions/%s", quota.Guid)
	return repo.gateway.UpdateResourceFromStruct(repo.config.ApiEndpoint(), path, quota)
}

func (repo CloudControllerSpaceQuotaRepository) AssociateSpaceWithQuota(spaceGuid string, quotaGuid string) error {
	path := fmt.Sprintf("/v2/space_quota_definitions/%s/spaces/%s", quotaGuid, spaceGuid)
	return repo.gateway.UpdateResource(repo.config.ApiEndpoint(), path, strings.NewReader(""))
}

func (repo CloudControllerSpaceQuotaRepository) UnassignQuotaFromSpace(spaceGuid string, quotaGuid string) error {
	path := fmt.Sprintf("/v2/space_quota_definitions/%s/spaces/%s", quotaGuid, spaceGuid)
	return repo.gateway.DeleteResource(repo.config.ApiEndpoint(), path)
}

func (repo CloudControllerSpaceQuotaRepository) Delete(quotaGuid string) (apiErr error) {
	path := fmt.Sprintf("/v2/space_quota_definitions/%s", quotaGuid)
	return repo.gateway.DeleteResource(repo.config.ApiEndpoint(), path)
}
