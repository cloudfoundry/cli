package api

import "cf/errors"

type FakeEndpointRepo struct {
	UpdateEndpointReceived string
	UpdateEndpointError    errors.Error
}

func (repo *FakeEndpointRepo) UpdateEndpoint(endpoint string) (string, errors.Error) {
	repo.UpdateEndpointReceived = endpoint
	return endpoint, repo.UpdateEndpointError
}
