package api

import (
	"cf"
	"cf/net"
	"cf/paginator"
	testpaginator "testhelpers/paginator"
)

type FakeOrgRepository struct {
	Organizations []cf.Organization

	CreateName string
	CreateOrgExists bool

	FindByNameName         string
	FindByNameErr          bool
	FindByNameNotFound     bool
	FindByNameOrganization cf.Organization

	RenameOrganization cf.Organization
	RenameNewName      string

	DeletedOrganization cf.Organization

	FindQuotaByNameName string
	FindQuotaByNameQuota cf.Quota
	FindQuotaByNameNotFound bool
	FindQuotaByNameErr bool

	UpdateQuotaOrg cf.Organization
	UpdateQuotaQuota cf.Quota
}

func (repo FakeOrgRepository) FindAll() (orgs []cf.Organization, apiResponse net.ApiResponse) {
	orgs = repo.Organizations
	return
}

func (repo FakeOrgRepository) Paginator() paginator.Paginator {
	results := []string{}
	for _, org := range repo.Organizations {
		results = append(results,org.Name)
	}
	return &testpaginator.FakePaginator{TotalResults:results}
}

func (repo *FakeOrgRepository) FindByName(name string) (org cf.Organization, apiResponse net.ApiResponse) {
	repo.FindByNameName = name
	org = repo.FindByNameOrganization

	if repo.FindByNameErr {
		apiResponse = net.NewApiResponseWithMessage("Error finding organization by name.")
	}

	if repo.FindByNameNotFound {
		apiResponse = net.NewNotFoundApiResponse("%s %s not found","Org", name)
	}

	return
}

func (repo *FakeOrgRepository) Create(name string) (apiResponse net.ApiResponse) {
	if repo.CreateOrgExists {
		apiResponse = net.NewApiResponse("Space already exists", cf.ORG_EXISTS, 400)
		return
	}
	repo.CreateName = name
	return
}

func (repo *FakeOrgRepository) Rename(org cf.Organization, newName string) (apiResponse net.ApiResponse) {
	repo.RenameOrganization = org
	repo.RenameNewName = newName
	return
}

func (repo *FakeOrgRepository) Delete(org cf.Organization) (apiResponse net.ApiResponse) {
	repo.DeletedOrganization = org
	return
}

func (repo *FakeOrgRepository) FindQuotaByName(name string) (quota cf.Quota, apiResponse net.ApiResponse) {
	repo.FindQuotaByNameName = name
	quota = repo.FindQuotaByNameQuota

	if repo.FindQuotaByNameNotFound {
		apiResponse = net.NewNotFoundApiResponse("%s %s not found","Org", name)
	}
	if repo.FindQuotaByNameErr {
		apiResponse = net.NewApiResponseWithMessage("Error finding quota")
	}

	return
}

func (repo *FakeOrgRepository) UpdateQuota(org cf.Organization, quota cf.Quota) (apiResponse net.ApiResponse) {
	repo.UpdateQuotaOrg = org
	repo.UpdateQuotaQuota = quota
	return
}
