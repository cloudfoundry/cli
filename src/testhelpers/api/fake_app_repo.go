package api

import (
	"cf"
	"cf/net"
)

type FakeApplicationRepository struct {
	ScaledApp cf.ApplicationFields

	StartAppGuid string
	StartAppErr     bool
	StartUpdatedApp cf.Application

	StopAppGuid  string
	StopAppErr     bool
	StopUpdatedApp cf.Application

	DeletedAppGuid string

	FindAllApps []cf.Application

	FindByNameName      string
	FindByNameApp       cf.Application
	FindByNameErr       bool
	FindByNameAuthErr   bool
	FindByNameNotFound  bool

	SetEnvAppGuid   string
	SetEnvVars  map[string]string
	SetEnvValue string
	SetEnvErr   bool

	CreateName string
	CreateBuildpackUrl string
	CreateStackGuid string
	CreateCommand string
	CreateMemory uint64
	CreateInstances int

	UpdatedApp cf.ApplicationFields
	UpdatedStackGuid string
	UpdateAppResult cf.Application

	RenameAppGuid     string
	RenameNewName string
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

func (repo *FakeApplicationRepository) SetEnv(appGuid string, envVars map[string]string) (apiResponse net.ApiResponse) {
	repo.SetEnvAppGuid = appGuid
	repo.SetEnvVars = envVars

	if repo.SetEnvErr {
		apiResponse = net.NewApiResponseWithMessage("Failed setting env")
	}
	return
}

func (repo *FakeApplicationRepository) Create(name, buildpackUrl, stackGuid, command string, memory uint64, instances int) (resultApp cf.Application, apiResponse net.ApiResponse) {
	repo.CreateName = name
	repo.CreateBuildpackUrl = buildpackUrl
	repo.CreateStackGuid = stackGuid
	repo.CreateCommand = command
	repo.CreateMemory = memory
	repo.CreateInstances = instances

	resultApp.Name = name
	resultApp.Guid = name+"-guid"
	resultApp.BuildpackUrl = buildpackUrl
	resultApp.Stack = cf.Stack{}
	resultApp.Stack.Guid = stackGuid
	resultApp.Command = command
	resultApp.Memory = memory
	resultApp.InstanceCount = instances

	return
}

func (repo *FakeApplicationRepository) Update(app cf.ApplicationFields, stackGuid string) (updatedApp cf.Application, apiResponse net.ApiResponse) {
	repo.UpdatedApp = app
	repo.UpdatedStackGuid = stackGuid
	updatedApp = repo.UpdateAppResult
	return
}

func (repo *FakeApplicationRepository) Delete(appGuid string) (apiResponse net.ApiResponse) {
	repo.DeletedAppGuid = appGuid
	return
}

func (repo *FakeApplicationRepository) Rename(appGuid, newName string) (apiResponse net.ApiResponse) {
	repo.RenameAppGuid = appGuid
	repo.RenameNewName = newName
	return
}

func (repo *FakeApplicationRepository) Scale(app cf.ApplicationFields) (apiResponse net.ApiResponse) {
	repo.ScaledApp = app
	return
}

func (repo *FakeApplicationRepository) Start(appGuid string) (updatedApp cf.Application, apiResponse net.ApiResponse) {
	repo.StartAppGuid = appGuid
	if repo.StartAppErr {
		apiResponse = net.NewApiResponseWithMessage("Error starting application")
	}
	updatedApp = repo.StartUpdatedApp
	return
}

func (repo *FakeApplicationRepository) Stop(appGuid string) (updatedApp cf.Application, apiResponse net.ApiResponse) {
	repo.StopAppGuid = appGuid
	if repo.StopAppErr {
		apiResponse = net.NewApiResponseWithMessage("Error stopping application")
	}
	updatedApp = repo.StopUpdatedApp
	return
}
