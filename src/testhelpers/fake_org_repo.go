package testhelpers

import (
	"cf"
	"cf/api"
)

type FakeOrgRepository struct {
	Organizations []cf.Organization

	CreateName string

	FindByNameName string
	FindByNameErr bool
	FindByNameOrganization cf.Organization
}

func (repo FakeOrgRepository) FindAll() (orgs []cf.Organization, apiErr *api.ApiError) {
	return repo.Organizations, nil
}

func (repo *FakeOrgRepository) FindByName(name string) (org cf.Organization, apiErr *api.ApiError) {
	repo.FindByNameName = name

	if repo.FindByNameErr {
		apiErr = api.NewApiErrorWithMessage("Error finding organization by name.")
	}
	return repo.FindByNameOrganization, apiErr
}

func (repo *FakeOrgRepository) Create(name string) (apiErr *api.ApiError) {
	repo.CreateName = name
	return
}
