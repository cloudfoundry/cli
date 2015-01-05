package fakes

import (
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
)

type FakeServiceBindingRepo struct {
	CreateServiceInstanceGuid string
	CreateApplicationGuid     string
	CreateErrorCode           string

	DeleteServiceInstance models.ServiceInstance
	DeleteApplicationGuid string
	DeleteBindingNotFound bool
	CreateNonHttpErrCode  string
}

func (repo *FakeServiceBindingRepo) Create(instanceGuid, appGuid string) (apiErr error) {
	repo.CreateServiceInstanceGuid = instanceGuid
	repo.CreateApplicationGuid = appGuid
	if repo.CreateNonHttpErrCode != "" {
		apiErr = errors.New(repo.CreateNonHttpErrCode)
		return
	}

	if repo.CreateErrorCode != "" {
		apiErr = errors.NewHttpError(400, repo.CreateErrorCode, "Error binding service")
	}

	return
}

func (repo *FakeServiceBindingRepo) Delete(instance models.ServiceInstance, appGuid string) (found bool, apiErr error) {
	repo.DeleteServiceInstance = instance
	repo.DeleteApplicationGuid = appGuid
	found = !repo.DeleteBindingNotFound
	return
}
