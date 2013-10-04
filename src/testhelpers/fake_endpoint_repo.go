package testhelpers

import (
	"cf/net"
)

type FakeEndpointRepo struct {
	UpdateEndpointEndpoint string
}

func (repo *FakeEndpointRepo) UpdateEndpoint(endpoint string) (apiStatus net.ApiStatus) {
	repo.UpdateEndpointEndpoint = endpoint
	return
}
