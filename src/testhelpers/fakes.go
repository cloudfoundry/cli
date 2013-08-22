package testhelpers

import (
	"cf/configuration"
	"cf"
	"errors"
)

type FakeOrgRepository struct {
	Organizations []cf.Organization

	OrganizationName string
	OrganizationByName cf.Organization
	OrganizationByNameErr bool
}

func (repo FakeOrgRepository) FindOrganizations(config *configuration.Configuration) (orgs []cf.Organization, err error) {
	return repo.Organizations, nil
}


func (repo *FakeOrgRepository) FindOrganizationByName(config *configuration.Configuration, name string) (org cf.Organization, err error) {
	repo.OrganizationName = name
	if repo.OrganizationByNameErr {
		err = errors.New("Error finding organization by name.")
	}
	return repo.OrganizationByName, err
}

func (repo FakeOrgRepository) OrganizationExists(config *configuration.Configuration, organization cf.Organization) (bool) {
	for _, o := range repo.Organizations{
		if o.Name == organization.Name || o.Guid == organization.Guid {
			return true
		}
	}
	return false
}
