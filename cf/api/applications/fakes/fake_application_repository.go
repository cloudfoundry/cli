package fakes

import (
	"errors"
	"sync"

	"github.com/cloudfoundry/cli/cf/api/applications"
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

	CreateRestageRequestArgs struct {
		AppGuid string
	}

	ReadFromSpaceStub        func(name string, spaceGuid string) (app models.Application, apiErr error)
	readFromSpaceMutex       sync.RWMutex
	readFromSpaceArgsForCall []struct {
		name      string
		spaceGuid string
	}
	readFromSpaceReturns struct {
		result1 models.Application
		result2 error
	}

	ReadEnvStub        func(guid string) (*models.Environment, error)
	readEnvMutex       sync.RWMutex
	readEnvArgsForCall []struct {
		guid string
	}
	readEnvReturns struct {
		result1 *models.Environment
		result2 error
	}
}

//counterfeiter section
func (fake *FakeApplicationRepository) ReadFromSpace(name string, spaceGuid string) (app models.Application, apiErr error) {
	fake.readFromSpaceMutex.Lock()
	defer fake.readFromSpaceMutex.Unlock()
	fake.readFromSpaceArgsForCall = append(fake.readFromSpaceArgsForCall, struct {
		name      string
		spaceGuid string
	}{name, spaceGuid})
	if fake.ReadFromSpaceStub != nil {
		return fake.ReadFromSpaceStub(name, spaceGuid)
	} else {
		return fake.readFromSpaceReturns.result1, fake.readFromSpaceReturns.result2
	}
}

func (fake *FakeApplicationRepository) ReadFromSpaceCallCount() int {
	fake.readFromSpaceMutex.RLock()
	defer fake.readFromSpaceMutex.RUnlock()
	return len(fake.readFromSpaceArgsForCall)
}

func (fake *FakeApplicationRepository) ReadFromSpaceArgsForCall(i int) (string, string) {
	fake.readFromSpaceMutex.RLock()
	defer fake.readFromSpaceMutex.RUnlock()
	return fake.readFromSpaceArgsForCall[i].name, fake.readFromSpaceArgsForCall[i].spaceGuid
}

func (fake *FakeApplicationRepository) ReadFromSpaceReturns(result1 models.Application, result2 error) {
	fake.readFromSpaceReturns = struct {
		result1 models.Application
		result2 error
	}{result1, result2}
}
func (fake *FakeApplicationRepository) ReadEnv(guid string) (*models.Environment, error) {
	fake.readEnvMutex.Lock()
	fake.readEnvArgsForCall = append(fake.readEnvArgsForCall, struct {
		guid string
	}{guid})
	fake.readEnvMutex.Unlock()
	if fake.ReadEnvStub != nil {
		return fake.ReadEnvStub(guid)
	} else {
		return fake.readEnvReturns.result1, fake.readEnvReturns.result2
	}
}

func (fake *FakeApplicationRepository) ReadEnvCallCount() int {
	fake.readEnvMutex.RLock()
	defer fake.readEnvMutex.RUnlock()
	return len(fake.readEnvArgsForCall)
}

func (fake *FakeApplicationRepository) ReadEnvArgsForCall(i int) string {
	fake.readEnvMutex.RLock()
	defer fake.readEnvMutex.RUnlock()
	return fake.readEnvArgsForCall[i].guid
}

func (fake *FakeApplicationRepository) ReadEnvReturns(result1 *models.Environment, result2 error) {
	fake.ReadEnvStub = nil
	fake.readEnvReturns = struct {
		result1 *models.Environment
		result2 error
	}{result1, result2}
}

//End counterfeiter section

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

func (repo *FakeApplicationRepository) CreateRestageRequest(guid string) (apiErr error) {
	repo.CreateRestageRequestArgs.AppGuid = guid
	return nil
}

var _ applications.ApplicationRepository = new(FakeApplicationRepository)
