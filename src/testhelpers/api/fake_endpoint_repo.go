package api

import (
	"cf/configuration"
	"cf/net"
	testconfig "testhelpers/configuration"
)

type FakeEndpointRepo struct {
	ConfigRepo testconfig.FakeConfigRepository
	Config     *configuration.Configuration

	UpdateEndpointReceived string
	UpdateEndpointError net.ApiResponse

	LoggregatorEndpointReturns struct {
		Endpoint string
		ApiResponse net.ApiResponse
	}

	UAAEndpointReturns struct {
		Endpoint string
		ApiResponse net.ApiResponse
	}
}

func (repo *FakeEndpointRepo) UpdateEndpoint(endpoint string) (finalEndpoint string, apiResponse net.ApiResponse) {
	repo.UpdateEndpointReceived = endpoint
	apiResponse = repo.UpdateEndpointError

	if apiResponse.IsNotSuccessful() {
		return
	}

	repo.Config, _ = repo.ConfigRepo.Get()
	repo.Config.Target = endpoint
	repo.ConfigRepo.Save()
	finalEndpoint = endpoint
	return
}

func (repo *FakeEndpointRepo) GetLoggregatorEndpoint() (endpoint string, apiResponse net.ApiResponse) {
	endpoint = repo.LoggregatorEndpointReturns.Endpoint
	apiResponse = repo.LoggregatorEndpointReturns.ApiResponse
	return
}

func (repo *FakeEndpointRepo) GetCloudControllerEndpoint() (endpoint string, apiResponse net.ApiResponse) {
	return
}

func (repo *FakeEndpointRepo) GetUAAEndpoint() (endpoint string, apiResponse net.ApiResponse) {
	endpoint = repo.UAAEndpointReturns.Endpoint
	apiResponse = repo.UAAEndpointReturns.ApiResponse
	return
}
