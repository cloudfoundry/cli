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

func (repo *FakeQuotaRepository) FindAll() (quotas []models.QuotaFields, apiErr errors.Error) {
	quotas = repo.FindAllQuotas

	return
}

func (repo *FakeQuotaRepository) FindByName(name string) (quota models.QuotaFields, apiErr errors.Error) {
	repo.FindByNameName = name
	quota = repo.FindByNameQuota

	if repo.FindByNameNotFound {
		apiErr = errors.NewNotFoundError("%s %s not found", "Org", name)
	}
	if repo.FindByNameErr {
		apiErr = errors.NewErrorWithMessage("Error finding quota")
	}

	return
}

func (repo *FakeQuotaRepository) Update(orgGuid, quotaGuid string) (apiErr errors.Error) {
	repo.UpdateOrgGuid = orgGuid
	repo.UpdateQuotaGuid = quotaGuid
	return
}
