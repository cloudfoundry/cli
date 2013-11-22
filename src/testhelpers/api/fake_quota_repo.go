package api

import (
	"cf"
	"cf/net"
)

type FakeQuotaRepository struct {
	FindAllQuotas []cf.QuotaFields

	FindByNameName     string
	FindByNameQuota    cf.QuotaFields
	FindByNameNotFound bool
	FindByNameErr      bool

	UpdateOrgGuid   string
	UpdateQuotaGuid string
}

func (repo *FakeQuotaRepository) FindAll() (quotas []cf.QuotaFields, apiResponse net.ApiResponse) {
	quotas = repo.FindAllQuotas

	return
}

func (repo *FakeQuotaRepository) FindByName(name string) (quota cf.QuotaFields, apiResponse net.ApiResponse) {
	repo.FindByNameName = name
	quota = repo.FindByNameQuota

	if repo.FindByNameNotFound {
		apiResponse = net.NewNotFoundApiResponse("%s %s not found", "Org", name)
	}
	if repo.FindByNameErr {
		apiResponse = net.NewApiResponseWithMessage("Error finding quota")
	}

	return
}

func (repo *FakeQuotaRepository) Update(orgGuid, quotaGuid string) (apiResponse net.ApiResponse) {
	repo.UpdateOrgGuid = orgGuid
	repo.UpdateQuotaGuid = quotaGuid
	return
}
