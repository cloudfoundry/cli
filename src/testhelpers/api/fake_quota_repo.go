package api

import (
	"cf"
	"cf/net"
)

type FakeQuotaRepository struct {
	FindByNameName string
	FindByNameQuota cf.Quota
	FindByNameNotFound bool
	FindByNameErr bool

	UpdateOrg cf.Organization
	UpdateQuota cf.Quota
}

func (repo *FakeQuotaRepository) FindByName(name string) (quota cf.Quota, apiResponse net.ApiResponse) {
	repo.FindByNameName = name
	quota = repo.FindByNameQuota

	if repo.FindByNameNotFound {
		apiResponse = net.NewNotFoundApiResponse("%s %s not found","Org", name)
	}
	if repo.FindByNameErr {
		apiResponse = net.NewApiResponseWithMessage("Error finding quota")
	}

	return
}

func (repo *FakeQuotaRepository) Update(org cf.Organization, quota cf.Quota) (apiResponse net.ApiResponse) {
	repo.UpdateOrg = org
	repo.UpdateQuota = quota
	return
}
