package api

import (
	"cf"
	"cf/errors"
	"cf/models"
)

type FakeOrgRepository struct {
	Organizations []models.Organization

	CreateName      string
	CreateOrgExists bool

	FindByNameName         string
	FindByNameErr          bool
	FindByNameNotFound     bool
	FindByNameOrganization models.Organization

	RenameOrganizationGuid string
	RenameNewName          string

	DeletedOrganizationGuid string
}

func (repo FakeOrgRepository) ListOrgs(cb func(models.Organization) bool) (apiResponse errors.Error) {
	for _, org := range repo.Organizations {
		if !cb(org) {
			break
		}
	}
	return
}

func (repo *FakeOrgRepository) FindByName(name string) (org models.Organization, apiResponse errors.Error) {
	repo.FindByNameName = name

	var foundOrg bool = false
	for _, anOrg := range repo.Organizations {
		if name == anOrg.Name {
			foundOrg = true
			org = anOrg
			break
		}
	}

	if repo.FindByNameErr || !foundOrg {
		apiResponse = errors.NewErrorWithMessage("Error finding organization by name.")
	}

	if repo.FindByNameNotFound {
		apiResponse = errors.NewNotFoundError("%s %s not found", "Org", name)
	}

	return
}

func (repo *FakeOrgRepository) Create(name string) (apiResponse errors.Error) {
	if repo.CreateOrgExists {
		apiResponse = errors.NewError("Space already exists", cf.ORG_EXISTS, 400)
		return
	}
	repo.CreateName = name
	return
}

func (repo *FakeOrgRepository) Rename(orgGuid string, name string) (apiResponse errors.Error) {
	repo.RenameOrganizationGuid = orgGuid
	repo.RenameNewName = name
	return
}

func (repo *FakeOrgRepository) Delete(orgGuid string) (apiResponse errors.Error) {
	repo.DeletedOrganizationGuid = orgGuid
	return
}
