package api

import (
	"cf/models"
	"cf/net"
	"net/http"
)

type FakeServiceBindingRepo struct {
	CreateServiceInstanceGuid string
	CreateApplicationGuid     string
	CreateErrorCode           string

	DeleteServiceInstance models.ServiceInstance
	DeleteApplicationGuid string
	DeleteBindingNotFound bool
}

func (repo *FakeServiceBindingRepo) Create(instanceGuid, appGuid string) (apiResponse net.ApiResponse) {
	repo.CreateServiceInstanceGuid = instanceGuid
	repo.CreateApplicationGuid = appGuid

	if repo.CreateErrorCode != "" {
		apiResponse = net.NewApiResponse("Error binding service", repo.CreateErrorCode, http.StatusBadRequest)
	}

	return
}

func (repo *FakeServiceBindingRepo) Delete(instance models.ServiceInstance, appGuid string) (found bool, apiResponse net.ApiResponse) {
	repo.DeleteServiceInstance = instance
	repo.DeleteApplicationGuid = appGuid
	found = !repo.DeleteBindingNotFound
	return
}
