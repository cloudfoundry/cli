package testhelpers

import (
	"cf"
	"cf/net"
)

type FakeOrgRepository struct {
	Organizations []cf.Organization

	CreateName string
	CreateOrgExists bool

	FindByNameName         string
	FindByNameErr          bool
	FindByNameOrganization cf.Organization
	DidNotFindOrganizationByName bool

	RenameOrganization cf.Organization
	RenameNewName      string

	DeletedOrganization cf.Organization
}

func (repo FakeOrgRepository) FindAll() (orgs []cf.Organization, apiErr *net.ApiError) {
	return repo.Organizations, nil
}

func (repo *FakeOrgRepository) FindByName(name string) (org cf.Organization, found bool, apiErr *net.ApiError) {
	found = true
	repo.FindByNameName = name

	if repo.DidNotFindOrganizationByName{
		found = false
		return
	}

	if repo.FindByNameErr {
		apiErr = net.NewApiErrorWithMessage("Error finding organization by name.")
	}

	org = repo.FindByNameOrganization
	return
}

func (repo *FakeOrgRepository) Create(name string) (apiErr *net.ApiError) {
	if repo.CreateOrgExists {
		apiErr = &net.ApiError{ErrorCode: net.ORG_EXISTS}
		return
	}
	repo.CreateName = name
	return
}

func (repo *FakeOrgRepository) Rename(org cf.Organization, newName string) (apiErr *net.ApiError) {
	repo.RenameOrganization = org
	repo.RenameNewName = newName
	return
}

func (repo *FakeOrgRepository) Delete(org cf.Organization) (apiErr *net.ApiError) {
	repo.DeletedOrganization = org
	return
}
