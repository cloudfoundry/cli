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

type FakeSpaceRepository struct {
	Spaces []cf.Space

	SpaceName string
	SpaceByName cf.Space
	SpaceByNameErr bool
}

func (repo FakeSpaceRepository) FindSpaces(config *configuration.Configuration) (spaces []cf.Space, err error) {
	return repo.Spaces, nil
}

func (repo *FakeSpaceRepository) FindSpaceByName(config *configuration.Configuration, name string) (space cf.Space, err error) {
	repo.SpaceName = name
	if repo.SpaceByNameErr {
		err = errors.New("Error finding space by name.")
	}
	return repo.SpaceByName, err
}
