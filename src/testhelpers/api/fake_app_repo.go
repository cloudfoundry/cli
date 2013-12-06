package api

import (
	"cf"
	"cf/net"
)

type FakeApplicationRepository struct {
	FindAllApps []cf.Application

	ReadName      string
	ReadApp       cf.Application
	ReadErr       bool
	ReadAuthErr   bool
	ReadNotFound  bool

	CreateName string
	CreateBuildpackUrl string
	CreateStackGuid string
	CreateCommand string
	CreateMemory uint64
	CreateInstances int

	UpdateParams cf.AppParams
	UpdateAppGuid string
	UpdateAppResult cf.Application
	UpdateErr       bool

	DeletedAppGuid string
}

func (repo *FakeApplicationRepository) Read(name string) (app cf.Application, apiResponse net.ApiResponse) {
	repo.ReadName = name
	app = repo.ReadApp

	if repo.ReadErr {
		apiResponse = net.NewApiResponseWithMessage("Error finding app by name.")
	}
	if repo.ReadAuthErr {
		apiResponse = net.NewApiResponse("Authentication failed.", "1000", 401)
	}
	if repo.ReadNotFound {
		apiResponse = net.NewNotFoundApiResponse("%s %s not found","App", name)
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

func (repo *FakeApplicationRepository) Update(appGuid string, params cf.AppParams) (updatedApp cf.Application, apiResponse net.ApiResponse) {
	repo.UpdateAppGuid = appGuid
	repo.UpdateParams = params
	updatedApp = repo.UpdateAppResult
	if repo.UpdateErr {
		apiResponse = net.NewApiResponseWithMessage("Error updating app.")
	}
	return
}

func (repo *FakeApplicationRepository) Delete(appGuid string) (apiResponse net.ApiResponse) {
	repo.DeletedAppGuid = appGuid
	return
}
