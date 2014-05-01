package api

import (
	"github.com/cloudfoundry/cli/cf/models"
)

type FakeQuotaRepository struct {
	FindAllReturns struct {
		Quotas []models.QuotaFields
		Error  error
	}

	FindByNameCalledWith struct {
		Name string
	}

	FindByNameReturns struct {
		Quota models.QuotaFields
		Error error
	}

	AssignQuotaToOrgCalledWith struct {
		OrgGuid   string
		QuotaGuid string
	}

	UpdateCalledWith struct {
		Name                  string
		MemoryLimit           uint64
		RoutesLimit           int
		ServicesLimit         int
		AllowPaidServicePlans bool
	}
	UpdateReturns struct {
		Error error
	}

	CreateCalledWith struct {
		Name                    string
		MemoryLimit             uint64
		RoutesLimit             int
		ServicesLimit           int
		NonBasicServicesAllowed bool
	}

	CreateReturns struct {
		Error error
	}

	DeleteCalledWith struct {
		Guid string
	}

	DeleteReturns struct {
		Error error
	}
}

func (repo *FakeQuotaRepository) FindAll() ([]models.QuotaFields, error) {
	return repo.FindAllReturns.Quotas, repo.FindAllReturns.Error
}

func (repo *FakeQuotaRepository) FindByName(name string) (models.QuotaFields, error) {
	repo.FindByNameCalledWith.Name = name
	return repo.FindByNameReturns.Quota, repo.FindByNameReturns.Error
}

func (repo *FakeQuotaRepository) AssignQuotaToOrg(orgGuid, quotaGuid string) (apiErr error) {
	repo.AssignQuotaToOrgCalledWith.OrgGuid = orgGuid
	repo.AssignQuotaToOrgCalledWith.QuotaGuid = quotaGuid
	return
}

func (repo *FakeQuotaRepository) Create(quota models.QuotaFields) error {
	repo.CreateCalledWith.Name = quota.Name
	repo.CreateCalledWith.MemoryLimit = quota.MemoryLimit
	repo.CreateCalledWith.RoutesLimit = quota.RoutesLimit
	repo.CreateCalledWith.ServicesLimit = quota.ServicesLimit
	repo.CreateCalledWith.NonBasicServicesAllowed = quota.NonBasicServicesAllowed

	return repo.CreateReturns.Error
}

func (repo *FakeQuotaRepository) Update(quota models.QuotaFields) error {
	repo.UpdateCalledWith.Name = quota.Name
	repo.UpdateCalledWith.MemoryLimit = quota.MemoryLimit
	repo.UpdateCalledWith.RoutesLimit = quota.RoutesLimit
	repo.UpdateCalledWith.ServicesLimit = quota.ServicesLimit
	repo.UpdateCalledWith.AllowPaidServicePlans = quota.NonBasicServicesAllowed

	return repo.UpdateReturns.Error
}

func (repo *FakeQuotaRepository) Delete(quotaGuid string) (apiErr error) {
	repo.DeleteCalledWith.Guid = quotaGuid

	return repo.DeleteReturns.Error
}
