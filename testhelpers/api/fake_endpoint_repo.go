package api

type FakeEndpointRepo struct {
	UpdateEndpointReceived string
	UpdateEndpointError    error
}

func (repo *FakeEndpointRepo) UpdateEndpoint(endpoint string) (string, error) {
	repo.UpdateEndpointReceived = endpoint
	return endpoint, repo.UpdateEndpointError
}
