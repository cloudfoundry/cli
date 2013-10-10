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

func (repo *FakeApplicationRepository) FindByName(name string) (app cf.Application, apiResponse net.ApiResponse) {
	repo.FindByNameName = name
	app = repo.FindByNameApp

	if repo.FindByNameErr {
		apiResponse = net.NewApiResponseWithMessage("Error finding app by name.")
	}
	if repo.FindByNameAuthErr {
		apiResponse = net.NewApiResponse("Authentication failed.", "1000", 401)
	}
	if repo.FindByNameNotFound {
		apiResponse = net.NewNotFoundApiResponse("%s %s not found","App", name)
	}

	return
}

func (repo *FakeApplicationRepository) SetEnv(app cf.Application, envVars map[string]string) (apiResponse net.ApiResponse) {
	repo.SetEnvApp = app
	repo.SetEnvVars = envVars

	if repo.SetEnvErr {
		apiResponse = net.NewApiResponseWithMessage("Failed setting env")
	}
	return
}

func (repo *FakeApplicationRepository) Create(newApp cf.Application) (createdApp cf.Application, apiResponse net.ApiResponse) {
	repo.CreatedApp = newApp

	createdApp = cf.Application{
		Name: newApp.Name,
		Guid: newApp.Name + "-guid",
	}

	return
}

func (repo *FakeApplicationRepository) Delete(app cf.Application) (apiResponse net.ApiResponse) {
	repo.DeletedApp = app
	return
}

func (repo *FakeApplicationRepository) Rename(app cf.Application, newName string) (apiResponse net.ApiResponse) {
	repo.RenameApp = app
	repo.RenameNewName = newName
	return
}

func (repo *FakeApplicationRepository) Scale(app cf.Application) (apiResponse net.ApiResponse) {
	repo.ScaledApp = app
	return
}

func (repo *FakeApplicationRepository) Start(app cf.Application) (updatedApp cf.Application, apiResponse net.ApiResponse) {
	repo.StartAppToStart = app
	if repo.StartAppErr {
		apiResponse = net.NewApiResponseWithMessage("Error starting application")
	}
	updatedApp = repo.StartUpdatedApp
	return
}

func (repo *FakeApplicationRepository) Stop(appToStop cf.Application) (updatedApp cf.Application, apiResponse net.ApiResponse) {
	repo.StopAppToStop = appToStop
	if repo.StopAppErr {
		apiResponse = net.NewApiResponseWithMessage("Error stopping application")
	}
	updatedApp = repo.StopUpdatedApp
	return
}

func (repo *FakeApplicationRepository) GetInstances(app cf.Application) (instances[]cf.ApplicationInstance, apiResponse net.ApiResponse) {
	time.Sleep(1*time.Millisecond) //needed for Windows only, otherwise it thinks error codes are not assigned
	errorCode := repo.GetInstancesErrorCodes[0]
	repo.GetInstancesErrorCodes = repo.GetInstancesErrorCodes[1:]

	instances = repo.GetInstancesResponses[0]
	repo.GetInstancesResponses = repo.GetInstancesResponses[1:]

	if errorCode != "" {
		apiResponse = net.NewApiResponse("Error staging app", errorCode, http.StatusBadRequest)
		return
	}

	return
}
