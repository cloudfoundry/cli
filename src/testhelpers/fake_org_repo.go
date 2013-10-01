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
	FindByNameNotFound     bool
	FindByNameOrganization cf.Organization

	RenameOrganization cf.Organization
	RenameNewName      string

	DeletedOrganization cf.Organization
}

func (repo FakeOrgRepository) FindAll() (orgs []cf.Organization, apiStatus net.ApiStatus) {
	orgs = repo.Organizations
	return
}

func (repo *FakeOrgRepository) FindByName(name string) (org cf.Organization, apiStatus net.ApiStatus) {
	repo.FindByNameName = name
	org = repo.FindByNameOrganization

	if repo.FindByNameErr {
		apiStatus = net.NewApiStatusWithMessage("Error finding organization by name.")
	}

	if repo.FindByNameNotFound {
		apiStatus = net.NewNotFoundApiStatus()
	}

	return
}

func (repo *FakeOrgRepository) Create(name string) (apiStatus net.ApiStatus) {
	if repo.CreateOrgExists {
		apiStatus = net.NewApiStatus("Space already exists", net.ORG_EXISTS, 400)
		return
	}
	repo.CreateName = name
	return
}

func (repo *FakeOrgRepository) Rename(org cf.Organization, newName string) (apiStatus net.ApiStatus) {
	repo.RenameOrganization = org
	repo.RenameNewName = newName
	return
}

func (repo *FakeOrgRepository) Delete(org cf.Organization) (apiStatus net.ApiStatus) {
	repo.DeletedOrganization = org
	return
}
