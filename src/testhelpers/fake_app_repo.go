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
	StartAppErr     bool
	StartUpdatedApp cf.Application

	StopAppToStop  cf.Application
	StopAppErr     bool
	StopUpdatedApp cf.Application

	DeletedApp cf.Application

	FindAllApps []cf.Application

	FindByNameName      string
	FindByNameApp       cf.Application
	FindByNameErr       bool
	FindByNameAuthErr   bool
	FindByNameNotFound  bool

	SetEnvApp   cf.Application
	SetEnvVars  map[string]string
	SetEnvValue string
	SetEnvErr   bool

	CreatedApp  cf.Application

	RenameApp     cf.Application
	RenameNewName string

	GetInstancesResponses  [][]cf.ApplicationInstance
	GetInstancesErrorCodes []string
}

func (repo *FakeApplicationRepository) FindByName(name string) (app cf.Application, apiStatus net.ApiStatus) {
	repo.FindByNameName = name
	app = repo.FindByNameApp

	if repo.FindByNameErr {
		apiStatus = net.NewApiStatusWithMessage("Error finding app by name.")
	}
	if repo.FindByNameAuthErr {
		apiStatus = net.NewApiStatus("Authentication failed.", "1000", 401)
	}
	if repo.FindByNameNotFound {
		apiStatus = net.NewNotFoundApiStatus()
	}

	return
}

func (repo *FakeApplicationRepository) SetEnv(app cf.Application, envVars map[string]string) (apiStatus net.ApiStatus) {
	repo.SetEnvApp = app
	repo.SetEnvVars = envVars

	if repo.SetEnvErr {
		apiStatus = net.NewApiStatusWithMessage("Failed setting env")
	}
	return
}

func (repo *FakeApplicationRepository) Create(newApp cf.Application) (createdApp cf.Application, apiStatus net.ApiStatus) {
	repo.CreatedApp = newApp

	createdApp = cf.Application{
		Name: newApp.Name,
		Guid: newApp.Name + "-guid",
	}

	return
}

func (repo *FakeApplicationRepository) Delete(app cf.Application) (apiStatus net.ApiStatus) {
	repo.DeletedApp = app
	return
}

func (repo *FakeApplicationRepository) Rename(app cf.Application, newName string) (apiStatus net.ApiStatus) {
	repo.RenameApp = app
	repo.RenameNewName = newName
	return
}

func (repo *FakeApplicationRepository) Scale(app cf.Application) (apiStatus net.ApiStatus) {
	repo.ScaledApp = app
	return
}

func (repo *FakeApplicationRepository) Start(app cf.Application) (updatedApp cf.Application, apiStatus net.ApiStatus) {
	repo.StartAppToStart = app
	if repo.StartAppErr {
		apiStatus = net.NewApiStatusWithMessage("Error starting application")
	}
	updatedApp = repo.StartUpdatedApp
	return
}

func (repo *FakeApplicationRepository) Stop(appToStop cf.Application) (updatedApp cf.Application, apiStatus net.ApiStatus) {
	repo.StopAppToStop = appToStop
	if repo.StopAppErr {
		apiStatus = net.NewApiStatusWithMessage("Error stopping application")
	}
	updatedApp = repo.StopUpdatedApp
	return
}

func (repo *FakeApplicationRepository) GetInstances(app cf.Application) (instances[]cf.ApplicationInstance, apiStatus net.ApiStatus) {
	time.Sleep(1*time.Millisecond) //needed for Windows only, otherwise it thinks error codes are not assigned
	errorCode := repo.GetInstancesErrorCodes[0]
	repo.GetInstancesErrorCodes = repo.GetInstancesErrorCodes[1:]

	instances = repo.GetInstancesResponses[0]
	repo.GetInstancesResponses = repo.GetInstancesResponses[1:]

	if errorCode != "" {
		apiStatus = net.NewApiStatus("Error staging app", errorCode, http.StatusBadRequest)
		return
	}

	return
}
