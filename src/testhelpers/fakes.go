package testhelpers

import (
	"cf/configuration"
	"cf"
	"errors"
)

type FakeAuthenticator struct {
	Config *configuration.Configuration
	Email string
	Password string

	AuthError bool
	AccessToken string
}

func (auth *FakeAuthenticator) Authenticate(config *configuration.Configuration, email string, password string) (err error) {
	auth.Config = config
	auth.Email = email
	auth.Password = password

	if auth.AccessToken == "" {
		auth.AccessToken = "BEARER some_access_token"
	}

	config.AccessToken = auth.AccessToken
	config.Save()

	if auth.AuthError {
		err = errors.New("Error authenticating.")
	}

	return
}

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

	CreatedApp cf.Application
	UploadedApp cf.Application
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

func (repo *FakeApplicationRepository) Create(config *configuration.Configuration, newApp cf.Application) (createdApp cf.Application, err error) {
	repo.CreatedApp = newApp

	createdApp = cf.Application{
		Name: newApp.Name,
		Guid: newApp.Name + "-guid",
	}

	return
}


func (repo *FakeApplicationRepository) Upload(config *configuration.Configuration, app cf.Application) (err error) {
	repo.UploadedApp = app

	return
}
