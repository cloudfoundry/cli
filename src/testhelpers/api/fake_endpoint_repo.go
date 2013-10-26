package api

import (
	"cf/net"
	"cf"
	testconfig "testhelpers/configuration"
	"cf/configuration"
)

type FakeEndpointRepo struct {
	ConfigRepo testconfig.FakeConfigRepository
	Config *configuration.Configuration

	UpdateEndpointEndpoint string
	UpdateEndpointError bool

	GetEndpointEndpoints map[cf.EndpointType]string
}

func (repo *FakeEndpointRepo) UpdateEndpoint(endpoint string) (finalEndpoint string, apiResponse net.ApiResponse) {
	repo.UpdateEndpointEndpoint = endpoint

	if repo.UpdateEndpointError {
		apiResponse = net.NewApiResponseWithMessage("Server error")
		return
	}
	repo.Config, _ = repo.ConfigRepo.Get()
	repo.Config.Target = endpoint
	repo.ConfigRepo.Save()
	finalEndpoint = endpoint
	return
}

func (repo *FakeEndpointRepo) GetEndpoint(name cf.EndpointType)  (endpoint string, apiResponse net.ApiResponse) {
	endpoint = repo.GetEndpointEndpoints[name]
	return
}
