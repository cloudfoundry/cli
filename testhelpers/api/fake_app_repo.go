package api

import (
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
)

type FakeApplicationRepository struct {
	FindAllApps []models.Application

	ReadArgs struct {
		Name string
	}
	ReadReturns struct {
		App   models.Application
		Error error
	}

	CreateAppParams []models.AppParams

	UpdateParams    models.AppParams
	UpdateAppGuid   string
	UpdateAppResult models.Application
	UpdateErr       bool

	DeletedAppGuid string
}

func (repo *FakeApplicationRepository) Read(name string) (app models.Application, apiErr error) {
	repo.ReadArgs.Name = name
	return repo.ReadReturns.App, repo.ReadReturns.Error
}

func (repo *FakeApplicationRepository) CreatedAppParams() (params models.AppParams) {
	if len(repo.CreateAppParams) > 0 {
		params = repo.CreateAppParams[0]
	}
	return
}

func (repo *FakeApplicationRepository) Create(params models.AppParams) (resultApp models.Application, apiErr error) {
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

func (repo *FakeApplicationRepository) Update(appGuid string, params models.AppParams) (updatedApp models.Application, apiErr error) {
	repo.UpdateAppGuid = appGuid
	repo.UpdateParams = params
	updatedApp = repo.UpdateAppResult
	if repo.UpdateErr {
		apiErr = errors.New("Error updating app.")
	}
	return
}

func (repo *FakeApplicationRepository) Delete(appGuid string) (apiErr error) {
	repo.DeletedAppGuid = appGuid
	return
}
