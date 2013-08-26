package testhelpers

import (
	"cf"
	"cf/configuration"
	"errors"
)

type FakeApplicationRepository struct {
	DeletedApp cf.Application

	FindAllApps []cf.Application

	AppName      string
	AppByName    cf.Application
	AppByNameErr bool

	SetEnvApp   cf.Application
	SetEnvName  string
	SetEnvValue string
	SetEnvErr   bool

	CreatedApp  cf.Application
	UploadedApp cf.Application
}

func (repo *FakeApplicationRepository) FindAll(config *configuration.Configuration) (apps []cf.Application, err error) {
	return repo.FindAllApps, err
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

func (repo *FakeApplicationRepository) Delete(config *configuration.Configuration, app cf.Application) (err error){
	repo.DeletedApp = app
	return
}


func (repo *FakeApplicationRepository) Upload(config *configuration.Configuration, app cf.Application) (err error) {
	repo.UploadedApp = app

	return
}
