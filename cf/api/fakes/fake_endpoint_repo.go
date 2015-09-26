package fakes

type FakeEndpointRepo struct {
	CallCount              int
	UpdateEndpointReceived string
	UpdateEndpointError    error
}

func (repo *FakeEndpointRepo) UpdateEndpoint(endpoint string) (string, error) {
	repo.CallCount += 1
	repo.UpdateEndpointReceived = endpoint
	return endpoint, repo.UpdateEndpointError
}
