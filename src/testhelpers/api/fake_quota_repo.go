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

func (repo *FakeQuotaRepository) FindAll() (quotas []models.QuotaFields, apiResponse errors.Error) {
	quotas = repo.FindAllQuotas

	return
}

func (repo *FakeQuotaRepository) FindByName(name string) (quota models.QuotaFields, apiResponse errors.Error) {
	repo.FindByNameName = name
	quota = repo.FindByNameQuota

	if repo.FindByNameNotFound {
		apiResponse = errors.NewNotFoundError("%s %s not found", "Org", name)
	}
	if repo.FindByNameErr {
		apiResponse = errors.NewErrorWithMessage("Error finding quota")
	}

	return
}

func (repo *FakeQuotaRepository) Update(orgGuid, quotaGuid string) (apiResponse errors.Error) {
	repo.UpdateOrgGuid = orgGuid
	repo.UpdateQuotaGuid = quotaGuid
	return
}
