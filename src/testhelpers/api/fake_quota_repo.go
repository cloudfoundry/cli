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

	CreateCalledWith struct {
		Name          string
		MemoryLimit   uint64
		RoutesLimit   uint
		ServicesLimit uint
	}

	CreateReturns struct {
		Error error
	}
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

func (repo *FakeQuotaRepository) Create(quota models.QuotaFields) error {
	repo.CreateCalledWith.Name = quota.Name
	repo.CreateCalledWith.MemoryLimit = quota.MemoryLimit
	repo.CreateCalledWith.RoutesLimit = quota.RoutesLimit
	repo.CreateCalledWith.ServicesLimit = quota.ServicesLimit

	return repo.CreateReturns.Error
}
