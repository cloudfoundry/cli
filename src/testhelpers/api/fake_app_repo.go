package api

import (
	"cf/errors"
	"cf/models"
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

func (repo *FakeApplicationRepository) Read(name string) (app models.Application, apiResponse errors.Error) {
	repo.ReadName = name
	app = repo.ReadApp

	if repo.ReadErr {
		apiResponse = errors.NewErrorWithMessage("Error finding app by name.")
	}
	if repo.ReadAuthErr {
		apiResponse = errors.NewHttpError(401, "", "", "1000", "Authentication failed.")
	}
	if repo.ReadNotFound {
		apiResponse = errors.NewNotFoundError("%s %s not found", "App", name)
	}

	return
}

func (repo *FakeApplicationRepository) CreatedAppParams() (params models.AppParams) {
	if len(repo.CreateAppParams) > 0 {
		params = repo.CreateAppParams[0]
	}
	return
}

func (repo *FakeApplicationRepository) Create(params models.AppParams) (resultApp models.Application, apiResponse errors.Error) {
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

func (repo *FakeApplicationRepository) Update(appGuid string, params models.AppParams) (updatedApp models.Application, apiResponse errors.Error) {
	repo.UpdateAppGuid = appGuid
	repo.UpdateParams = params
	updatedApp = repo.UpdateAppResult
	if repo.UpdateErr {
		apiResponse = errors.NewErrorWithMessage("Error updating app.")
	}
	return
}

func (repo *FakeApplicationRepository) Delete(appGuid string) (apiResponse errors.Error) {
	repo.DeletedAppGuid = appGuid
	return
}
