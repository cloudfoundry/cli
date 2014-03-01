package api

import (
	"cf/errors"
	"cf/models"
)

type FakeServiceBindingRepo struct {
	CreateServiceInstanceGuid string
	CreateApplicationGuid     string
	CreateErrorCode           string

	DeleteServiceInstance models.ServiceInstance
	DeleteApplicationGuid string
	DeleteBindingNotFound bool
}

func (repo *FakeServiceBindingRepo) Create(instanceGuid, appGuid string) (apiErr errors.Error) {
	repo.CreateServiceInstanceGuid = instanceGuid
	repo.CreateApplicationGuid = appGuid

	if repo.CreateErrorCode != "" {
		apiErr = errors.NewError("Error binding service", repo.CreateErrorCode)
	}

	return
}

func (repo *FakeServiceBindingRepo) Delete(instance models.ServiceInstance, appGuid string) (found bool, apiErr errors.Error) {
	repo.DeleteServiceInstance = instance
	repo.DeleteApplicationGuid = appGuid
	found = !repo.DeleteBindingNotFound
	return
}
