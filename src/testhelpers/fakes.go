package testhelpers

import (
	"cf/api"
	"cf/configuration"
)

type FakeOrgRepository struct {
	Organizations []api.Organization
}

func (repo *FakeOrgRepository) FindOrganizations(config *configuration.Configuration) (orgs []api.Organization, err error) {
	return repo.Organizations, nil
}
