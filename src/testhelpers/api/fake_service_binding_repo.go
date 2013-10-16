package api

import (
	"cf"
	"cf/net"
	"net/http"
)

type FakeServiceBindingRepo struct {
	CreateServiceInstance cf.ServiceInstance
	CreateApplication cf.Application
	CreateErrorCode string

	DeleteServiceInstance cf.ServiceInstance
	DeleteApplication cf.Application
	DeleteBindingNotFound bool
}

func (repo *FakeServiceBindingRepo) Create(instance cf.ServiceInstance, app cf.Application) (apiResponse net.ApiResponse) {
	repo.CreateServiceInstance = instance
	repo.CreateApplication = app

	if repo.CreateErrorCode != "" {
		apiResponse = net.NewApiResponse("Error binding service", repo.CreateErrorCode, http.StatusBadRequest)
	}

	return
}

func (repo *FakeServiceBindingRepo) Delete(instance cf.ServiceInstance, app cf.Application) (found bool, apiResponse net.ApiResponse) {
	repo.DeleteServiceInstance = instance
	repo.DeleteApplication = app
	found = !repo.DeleteBindingNotFound
	return
}
