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

type FakeApplicationRepository struct {
	AppName string
	AppByName cf.Application
	AppByNameErr bool

	SetEnvApp cf.Application
	SetEnvName string
	SetEnvValue string
	SetEnvErr bool
}

func (repo *FakeApplicationRepository) FindByName(config *configuration.Configuration, name string) (app cf.Application, err error) {
	repo.AppName = name
	if repo.AppByNameErr {
		err = errors.New("Error finding app by name.")
	}
	return repo.AppByName, err
}

func (repo *FakeApplicationRepository) SetEnv(config *configuration.Configuration, app cf.Application, name string, value string) (err error) {
	repo.SetEnvApp = app
	repo.SetEnvName = name
	repo.SetEnvValue = value

	if repo.SetEnvErr {
		err = errors.New("Error setting env.")
	}
	return
}
