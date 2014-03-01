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
		Endpoint string
		Error    errors.Error
	}

	UAAEndpointReturns struct {
		Endpoint string
		Error    errors.Error
	}
}

func (repo *FakeEndpointRepo) UpdateEndpoint(endpoint string) (finalEndpoint string, apiErr errors.Error) {
	repo.UpdateEndpointReceived = endpoint
	apiErr = repo.UpdateEndpointError

	if apiErr != nil {
		return
	}

	repo.Config.SetApiEndpoint(endpoint)
	finalEndpoint = endpoint
	return
}

func (repo *FakeEndpointRepo) GetLoggregatorEndpoint() (endpoint string, apiErr errors.Error) {
	endpoint = repo.LoggregatorEndpointReturns.Endpoint
	apiErr = repo.LoggregatorEndpointReturns.Error
	return
}

func (repo *FakeEndpointRepo) GetCloudControllerEndpoint() (endpoint string, apiErr errors.Error) {
	return
}

func (repo *FakeEndpointRepo) GetUAAEndpoint() (endpoint string, apiErr errors.Error) {
	endpoint = repo.UAAEndpointReturns.Endpoint
	apiErr = repo.UAAEndpointReturns.Error
	return
}
