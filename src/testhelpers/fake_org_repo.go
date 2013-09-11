package testhelpers

import (
	"cf"
	"cf/api"
)

type FakeOrgRepository struct {
	Organizations []cf.Organization

	OrganizationName string
	OrganizationByName cf.Organization
	OrganizationByNameErr bool
}

func (repo *FakeOrgRepository) CreateOrgRepository(name string) (apiErr *api.ApiError) {
	repo.OrganizationName = name
	return
}

func (repo FakeOrgRepository) FindAll() (orgs []cf.Organization, apiErr *api.ApiError) {
	return repo.Organizations, nil
}

func (repo *FakeOrgRepository) FindByName(name string) (org cf.Organization, apiErr *api.ApiError) {
	repo.OrganizationName = name
	if repo.OrganizationByNameErr {
		apiErr = api.NewApiErrorWithMessage("Error finding organization by name.")
	}
	return repo.OrganizationByName, apiErr
}

