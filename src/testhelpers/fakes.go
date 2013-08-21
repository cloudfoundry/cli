package testhelpers

import (
	"cf/api"
	"cf/configuration"
)

type FakeOrgRepository struct {
	Organizations []api.Organization
}

func (repo FakeOrgRepository) FindOrganizations(config *configuration.Configuration) (orgs []api.Organization, err error) {
	return repo.Organizations, nil
}

func (repo FakeOrgRepository) OrganizationExists(config *configuration.Configuration, organization api.Organization) (bool) {
	for _, o := range repo.Organizations{
		if o.Name == organization.Name || o.Guid == organization.Guid{
			return true
		}
	}
	return false
}
