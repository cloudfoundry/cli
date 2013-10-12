package api

import (
	"cf/net"
	"cf"
)

type FakeEndpointRepo struct {
	UpdateEndpointEndpoint string

	GetEndpointEndpoints map[cf.EndpointType]string
}

func (repo *FakeEndpointRepo) UpdateEndpoint(endpoint string) (apiResponse net.ApiResponse) {
	repo.UpdateEndpointEndpoint = endpoint
	return
}

func (repo *FakeEndpointRepo) GetEndpoint(name cf.EndpointType)  (endpoint string, apiResponse net.ApiResponse) {
	endpoint = repo.GetEndpointEndpoints[name]
	return
}
