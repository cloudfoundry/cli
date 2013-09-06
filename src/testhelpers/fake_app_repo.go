package testhelpers

import (
	"cf"
	"cf/configuration"
	"errors"
	"bytes"
)

type FakeApplicationRepository struct {
	StartedApp cf.Application
	StartAppErr bool

	StoppedApp cf.Application
	StopAppErr bool

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
	UploadedZipBuffer *bytes.Buffer

	GetInstancesResponses [][]cf.ApplicationInstance
	GetInstancesErrorCodes []int
}

func (repo *FakeApplicationRepository) FindByName(name string) (app cf.Application, err error) {
	repo.AppName = name
	if repo.AppByNameErr {
		err = errors.New("Error finding app by name.")
	}
	return repo.AppByName, err
}

func (repo *FakeApplicationRepository) SetEnv(app cf.Application, name string, value string) (err error) {
	repo.SetEnvApp = app
	repo.SetEnvName = name
	repo.SetEnvValue = value

	if repo.SetEnvErr {
		err = errors.New("Error setting env.")
	}
	return
}

func (repo *FakeApplicationRepository) Create(newApp cf.Application) (createdApp cf.Application, err error) {
	repo.CreatedApp = newApp

	createdApp = cf.Application{
		Name: newApp.Name,
		Guid: newApp.Name + "-guid",
	}

	return
}

func (repo *FakeApplicationRepository) Delete(app cf.Application) (err error){
	repo.DeletedApp = app
	return
}


func (repo *FakeApplicationRepository) Upload(app cf.Application, zipBuffer *bytes.Buffer) (err error) {
	repo.UploadedZipBuffer = zipBuffer
	repo.UploadedApp = app

	return
}

func (repo *FakeApplicationRepository) Start(app cf.Application) (err error){
	repo.StartedApp = app
	if repo.StartAppErr {
		err = errors.New("Error starting app.")
	}
	return
}

func (repo *FakeApplicationRepository) Stop(app cf.Application) (err error){
	repo.StoppedApp = app
	if repo.StopAppErr {
		err = errors.New("Error stopping app.")
	}
	return
}

func (repo *FakeApplicationRepository) GetInstances(config *configuration.Configuration, app cf.Application) (instances[]cf.ApplicationInstance, errorCode int, err error) {
	errorCode = repo.GetInstancesErrorCodes[0]
	repo.GetInstancesErrorCodes = repo.GetInstancesErrorCodes[1:]


	instances = repo.GetInstancesResponses[0]
	repo.GetInstancesResponses = repo.GetInstancesResponses[1:]

	if errorCode != 0 {
		err = errors.New("Error while starting app")
		return
	}

	return
}
