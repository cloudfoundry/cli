package testhelpers

import (
	"cf"
	"cf/net"
)

type FakeOrgRepository struct {
	Organizations []cf.Organization

	CreateName string

	FindByNameName string
	FindByNameErr bool
	FindByNameOrganization cf.Organization

	RenameOrganization cf.Organization
	RenameNewName string

	DeletedOrganization cf.Organization
}

func (repo FakeOrgRepository) FindAll() (orgs []cf.Organization, apiErr *net.ApiError) {
	return repo.Organizations, nil
}

func (repo *FakeOrgRepository) FindByName(name string) (org cf.Organization, apiErr *net.ApiError) {
	repo.FindByNameName = name

	if repo.FindByNameErr {
		apiErr = net.NewApiErrorWithMessage("Error finding organization by name.")
	}
	return repo.FindByNameOrganization, apiErr
}

func (repo *FakeOrgRepository) Create(name string) (apiErr *net.ApiError) {
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
