package api

import (
	"cf/configuration"
	"cf/errors"
)

type FakeEndpointRepo struct {
	Config configuration.ReadWriter

	UpdateEndpointReceived string
	UpdateEndpointError    errors.Error

	LoggregatorEndpointReturns struct {
		Endpoint    string
		ApiResponse errors.Error
	}

	UAAEndpointReturns struct {
		Endpoint    string
		ApiResponse errors.Error
	}
}

func (repo *FakeEndpointRepo) UpdateEndpoint(endpoint string) (finalEndpoint string, apiResponse errors.Error) {
	repo.UpdateEndpointReceived = endpoint
	apiResponse = repo.UpdateEndpointError

	if apiResponse != nil {
		return
	}

	repo.Config.SetApiEndpoint(endpoint)
	finalEndpoint = endpoint
	return
}

func (repo *FakeEndpointRepo) GetLoggregatorEndpoint() (endpoint string, apiResponse errors.Error) {
	endpoint = repo.LoggregatorEndpointReturns.Endpoint
	apiResponse = repo.LoggregatorEndpointReturns.ApiResponse
	return
}

func (repo *FakeEndpointRepo) GetCloudControllerEndpoint() (endpoint string, apiResponse errors.Error) {
	return
}

func (repo *FakeEndpointRepo) GetUAAEndpoint() (endpoint string, apiResponse errors.Error) {
	endpoint = repo.UAAEndpointReturns.Endpoint
	apiResponse = repo.UAAEndpointReturns.ApiResponse
	return
}
