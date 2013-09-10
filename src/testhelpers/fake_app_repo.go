package testhelpers

import (
	"cf"
	"bytes"
	"cf/api"
	"net/http"
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

func (repo *FakeApplicationRepository) FindByName(name string) (app cf.Application, apiErr *api.ApiError) {
	repo.AppName = name
	if repo.AppByNameErr {
		apiErr = api.NewApiErrorWithMessage("Error finding app by name.")
	}
	return repo.AppByName, apiErr
}

func (repo *FakeApplicationRepository) SetEnv(app cf.Application, name string, value string) (apiErr *api.ApiError) {
	repo.SetEnvApp = app
	repo.SetEnvName = name
	repo.SetEnvValue = value

	if repo.SetEnvErr {
		apiErr = api.NewApiErrorWithMessage("Error setting env.")
	}
	return
}

func (repo *FakeApplicationRepository) Create(newApp cf.Application) (createdApp cf.Application, apiErr *api.ApiError) {
	repo.CreatedApp = newApp

	createdApp = cf.Application{
		Name: newApp.Name,
		Guid: newApp.Name + "-guid",
	}

	return
}

func (repo *FakeApplicationRepository) Delete(app cf.Application) (apiErr *api.ApiError){
	repo.DeletedApp = app
	return
}


func (repo *FakeApplicationRepository) Upload(app cf.Application, zipBuffer *bytes.Buffer) (apiErr *api.ApiError) {
	repo.UploadedZipBuffer = zipBuffer
	repo.UploadedApp = app

	return
}

func (repo *FakeApplicationRepository) Start(app cf.Application) (apiErr *api.ApiError){
	repo.StartedApp = app
	if repo.StartAppErr {
		apiErr = api.NewApiErrorWithMessage("Error starting app.")
	}
	return
}

func (repo *FakeApplicationRepository) Stop(app cf.Application) (apiErr *api.ApiError){
	repo.StoppedApp = app
	if repo.StopAppErr {
		apiErr = api.NewApiErrorWithMessage("Error stopping app.")
	}
	return
}

func (repo *FakeApplicationRepository) GetInstances(app cf.Application) (instances[]cf.ApplicationInstance, apiErr *api.ApiError) {
	errorCode := repo.GetInstancesErrorCodes[0]
	repo.GetInstancesErrorCodes = repo.GetInstancesErrorCodes[1:]

	instances = repo.GetInstancesResponses[0]
	repo.GetInstancesResponses = repo.GetInstancesResponses[1:]

	if errorCode != 0 {
		apiErr = api.NewApiError("Error while starting app", errorCode, http.StatusBadRequest)
		return
	}

	return
}
