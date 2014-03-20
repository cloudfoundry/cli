package api

import (
	"cf/errors"
	"cf/models"
)

type FakeQuotaRepository struct {
	FindAllQuotas []models.QuotaFields

	FindByNameName     string
	FindByNameQuota    models.QuotaFields
	FindByNameNotFound bool
	FindByNameErr      bool

	UpdateOrgGuid   string
	UpdateQuotaGuid string
}

func (repo *FakeQuotaRepository) FindAll() (quotas []models.QuotaFields, apiErr error) {
	quotas = repo.FindAllQuotas

	return
}

func (repo *FakeQuotaRepository) FindByName(name string) (quota models.QuotaFields, apiErr error) {
	repo.FindByNameName = name
	quota = repo.FindByNameQuota

	if repo.FindByNameNotFound {
		apiErr = errors.NewModelNotFoundError("Org", name)
	}
	if repo.FindByNameErr {
		apiErr = errors.New("Error finding quota")
	}

	return
}

func (repo *FakeQuotaRepository) Update(orgGuid, quotaGuid string) (apiErr error) {
	repo.UpdateOrgGuid = orgGuid
	repo.UpdateQuotaGuid = quotaGuid
	return
}
