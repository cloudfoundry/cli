package api

import (
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
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

func (repo FakeOrgRepository) ListOrgs(cb func(models.Organization) bool) (apiErr error) {
	for _, org := range repo.Organizations {
		if !cb(org) {
			break
		}
	}
	return
}

func (repo *FakeOrgRepository) FindByName(name string) (org models.Organization, apiErr error) {
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
		apiErr = errors.New("Error finding organization by name.")
	}

	if repo.FindByNameNotFound {
		apiErr = errors.NewModelNotFoundError("Org", name)
	}

	return
}

func (repo *FakeOrgRepository) Create(name string) (apiErr error) {
	if repo.CreateOrgExists {
		apiErr = errors.NewHttpError(400, errors.ORG_EXISTS, "Space already exists")
		return
	}
	repo.CreateName = name
	return
}

func (repo *FakeOrgRepository) Rename(orgGuid string, name string) (apiErr error) {
	repo.RenameOrganizationGuid = orgGuid
	repo.RenameNewName = name
	return
}

func (repo *FakeOrgRepository) Delete(orgGuid string) (apiErr error) {
	repo.DeletedOrganizationGuid = orgGuid
	return
}
