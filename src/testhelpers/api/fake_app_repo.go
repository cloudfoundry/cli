package api

import (
	"cf/models"
	"cf/net"
)

type FakeApplicationRepository struct {
	FindAllApps []models.Application

	ReadName     string
	ReadApp      models.Application
	ReadErr      bool
	ReadAuthErr  bool
	ReadNotFound bool

	CreateAppParams []models.AppParams

	UpdateParams    models.AppParams
	UpdateAppGuid   string
	UpdateAppResult models.Application
	UpdateErr       bool

	DeletedAppGuid string
}

func (repo *FakeApplicationRepository) Read(name string) (app models.Application, apiResponse net.ApiResponse) {
	repo.ReadName = name
	app = repo.ReadApp

	if repo.ReadErr {
		apiResponse = net.NewApiResponseWithMessage("Error finding app by name.")
	}
	if repo.ReadAuthErr {
		apiResponse = net.NewApiResponse("Authentication failed.", "1000", 401)
	}
	if repo.ReadNotFound {
		apiResponse = net.NewNotFoundApiResponse("%s %s not found", "App", name)
	}

	return
}

func (repo *FakeApplicationRepository) CreatedAppParams() (params models.AppParams) {
	if len(repo.CreateAppParams) > 0 {
		params = repo.CreateAppParams[0]
	}
	return
}

func (repo *FakeApplicationRepository) Create(params models.AppParams) (resultApp models.Application, apiResponse net.ApiResponse) {
	if repo.CreateAppParams == nil {
		repo.CreateAppParams = []models.AppParams{}
	}

	repo.CreateAppParams = append(repo.CreateAppParams, params)

	resultApp.Guid = *params.Name + "-guid"
	resultApp.Name = *params.Name
	resultApp.State = "stopped"
	resultApp.EnvironmentVars = map[string]string{}

	if params.SpaceGuid != nil {
		resultApp.SpaceGuid = *params.SpaceGuid
	}
	if params.BuildpackUrl != nil {
		resultApp.BuildpackUrl = *params.BuildpackUrl
	}
	if params.Command != nil {
		resultApp.Command = *params.Command
	}
	if params.DiskQuota != nil {
		resultApp.DiskQuota = *params.DiskQuota
	}
	if params.InstanceCount != nil {
		resultApp.InstanceCount = *params.InstanceCount
	}
	if params.Memory != nil {
		resultApp.Memory = *params.Memory
	}
	if params.EnvironmentVars != nil {
		resultApp.EnvironmentVars = *params.EnvironmentVars
	}

	return
}

func (repo *FakeApplicationRepository) Update(appGuid string, params models.AppParams) (updatedApp models.Application, apiResponse net.ApiResponse) {
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
