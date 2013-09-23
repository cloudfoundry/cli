package testhelpers

import (
	"cf"
	"cf/net"
	"net/http"
	"time"
)

type FakeApplicationRepository struct {

	ScaledApp cf.Application

	StartAppToStart cf.Application
	StartAppErr bool
	StartUpdatedApp cf.Application

	StopAppToStop cf.Application
	StopAppErr bool
	StopUpdatedApp cf.Application

	DeletedApp cf.Application

	FindAllApps []cf.Application

	AppName      string
	AppByName    cf.Application
	AppByNameErr bool
	AppByNameAuthErr bool

	SetEnvApp   cf.Application
	SetEnvVars  map[string]string
	SetEnvValue string
	SetEnvErr   bool

	CreatedApp  cf.Application

	RenameApp cf.Application
	RenameNewName string

	GetInstancesResponses [][]cf.ApplicationInstance
	GetInstancesErrorCodes []string
}

func (repo *FakeApplicationRepository) FindByName(name string) (app cf.Application, apiErr *net.ApiError) {
	repo.AppName = name
	if repo.AppByNameErr {
		apiErr = net.NewApiErrorWithMessage("Error finding app by name.")
	}
	if repo.AppByNameAuthErr {
		apiErr = net.NewApiError("Authentication failed.", "1000", 401)
	}
	return repo.AppByName, apiErr
}

func (repo *FakeApplicationRepository) SetEnv(app cf.Application, envVars map[string]string) (apiErr *net.ApiError) {
	repo.SetEnvApp = app
	repo.SetEnvVars= envVars

	if repo.SetEnvErr {
		apiErr = net.NewApiErrorWithMessage("Failed setting env")
	}
	return
}

func (repo *FakeApplicationRepository) Create(newApp cf.Application) (createdApp cf.Application, apiErr *net.ApiError) {
	repo.CreatedApp = newApp

	createdApp = cf.Application{
		Name: newApp.Name,
		Guid: newApp.Name + "-guid",
	}

	return
}

func (repo *FakeApplicationRepository) Delete(app cf.Application) (apiErr *net.ApiError){
	repo.DeletedApp = app
	return
}

func (repo *FakeApplicationRepository) Rename(app cf.Application, newName string) (apiErr *net.ApiError) {
	repo.RenameApp = app
	repo.RenameNewName = newName
	return
}

func (repo *FakeApplicationRepository) Scale(app cf.Application) (apiErr *net.ApiError){
	repo.ScaledApp = app
	return
}

func (repo *FakeApplicationRepository) Start(app cf.Application) (updatedApp cf.Application, apiErr *net.ApiError){
	repo.StartAppToStart = app
	if repo.StartAppErr {
		apiErr = net.NewApiErrorWithMessage("Error starting application")
	}
	updatedApp = repo.StartUpdatedApp
	return
}

func (repo *FakeApplicationRepository) Stop(appToStop cf.Application) (updatedApp cf.Application, apiErr *net.ApiError){
	repo.StopAppToStop = appToStop
	if repo.StopAppErr {
		apiErr = net.NewApiErrorWithMessage("Error stopping application")
	}
	updatedApp = repo.StopUpdatedApp
	return
}

func (repo *FakeApplicationRepository) GetInstances(app cf.Application) (instances[]cf.ApplicationInstance, apiErr *net.ApiError) {
	time.Sleep(1 * time.Millisecond) //needed for Windows only, otherwise it thinks error codes are not assigned
	errorCode := repo.GetInstancesErrorCodes[0]
	repo.GetInstancesErrorCodes = repo.GetInstancesErrorCodes[1:]

	instances = repo.GetInstancesResponses[0]
	repo.GetInstancesResponses = repo.GetInstancesResponses[1:]

	if errorCode != "" {
		apiErr = net.NewApiError("Error staging app", errorCode, http.StatusBadRequest)
		return
	}

	return
}
