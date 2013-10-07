package testhelpers

import (
	"cf/net"
)

type FakeEndpointRepo struct {
	UpdateEndpointEndpoint string
}

func (repo *FakeEndpointRepo) UpdateEndpoint(endpoint string) (apiResponse net.ApiResponse) {
	repo.UpdateEndpointEndpoint = endpoint
	return
}
