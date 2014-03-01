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

func (repo FakeOrgRepository) ListOrgs(cb func(models.Organization) bool) (apiErr errors.Error) {
	for _, org := range repo.Organizations {
		if !cb(org) {
			break
		}
	}
	return
}

func (repo *FakeOrgRepository) FindByName(name string) (org models.Organization, apiErr errors.Error) {
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
		apiErr = errors.NewErrorWithMessage("Error finding organization by name.")
	}

	if repo.FindByNameNotFound {
		apiErr = errors.NewNotFoundError("%s %s not found", "Org", name)
	}

	return
}

func (repo *FakeOrgRepository) Create(name string) (apiErr errors.Error) {
	if repo.CreateOrgExists {
		apiErr = errors.NewError("Space already exists", cf.ORG_EXISTS)
		return
	}
	repo.CreateName = name
	return
}

func (repo *FakeOrgRepository) Rename(orgGuid string, name string) (apiErr errors.Error) {
	repo.RenameOrganizationGuid = orgGuid
	repo.RenameNewName = name
	return
}

func (repo *FakeOrgRepository) Delete(orgGuid string) (apiErr errors.Error) {
	repo.DeletedOrganizationGuid = orgGuid
	return
}
