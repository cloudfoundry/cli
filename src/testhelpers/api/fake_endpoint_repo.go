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
	GetEndpointEndpoints map[cf.EndpointType]string
}

func (repo *FakeEndpointRepo) UpdateEndpoint(endpoint string) (apiResponse net.ApiResponse) {
	repo.UpdateEndpointEndpoint = endpoint

	repo.Config, _ = repo.ConfigRepo.Get()
	repo.Config.Target = endpoint
	repo.ConfigRepo.Save()

	return
}

func (repo *FakeEndpointRepo) GetEndpoint(name cf.EndpointType)  (endpoint string, apiResponse net.ApiResponse) {
	endpoint = repo.GetEndpointEndpoints[name]
	return
}
