package api

import (
	"cf/api/resources"
	"cf/configuration"
	"cf/errors"
	"cf/models"
	"cf/net"
	"fmt"
	"net/url"
	"strings"
)

type QuotaRepository interface {
	FindAll() (quotas []models.QuotaFields, apiErr error)
	FindByName(name string) (quota models.QuotaFields, apiErr error)
	Update(orgGuid, quotaGuid string) (apiErr error)
}

type CloudControllerQuotaRepository struct {
	config  configuration.Reader
	gateway net.Gateway
}

func NewCloudControllerQuotaRepository(config configuration.Reader, gateway net.Gateway) (repo CloudControllerQuotaRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerQuotaRepository) findAllWithPath(path string) ([]models.QuotaFields, error) {
	var quotas []models.QuotaFields
	apiErr := repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		path,
		resources.QuotaResource{},
		func(resource interface{}) bool {
			if qr, ok := resource.(resources.QuotaResource); ok {
				quotas = append(quotas, qr.ToFields())
			}
			return true
		})
	return quotas, apiErr
}

func (repo CloudControllerQuotaRepository) FindAll() (quotas []models.QuotaFields, apiErr error) {
	return repo.findAllWithPath("/v2/quota_definitions")
}

func (repo CloudControllerQuotaRepository) FindByName(name string) (quota models.QuotaFields, apiErr error) {
	path := fmt.Sprintf("/v2/quota_definitions?q=%s", url.QueryEscape("name:"+name))
	quotas, apiErr := repo.findAllWithPath(path)
	if apiErr != nil {
		return
	}

	if len(quotas) == 0 {
		apiErr = errors.NewModelNotFoundError("Quota", name)
		return
	}

	quota = quotas[0]
	return
}

func (repo CloudControllerQuotaRepository) Update(orgGuid, quotaGuid string) (apiErr error) {
	path := fmt.Sprintf("%s/v2/organizations/%s", repo.config.ApiEndpoint(), orgGuid)
	data := fmt.Sprintf(`{"quota_definition_guid":"%s"}`, quotaGuid)
	return repo.gateway.UpdateResource(path, strings.NewReader(data))
}
