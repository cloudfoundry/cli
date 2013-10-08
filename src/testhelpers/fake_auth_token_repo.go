package testhelpers

import (
	"cf"
	"cf/net"
)

type FakeAuthTokenRepo struct {
	CreatedServiceAuthToken cf.ServiceAuthToken
}

func (repo *FakeAuthTokenRepo) Create(authToken cf.ServiceAuthToken) (apiResponse net.ApiResponse) {
	repo.CreatedServiceAuthToken = authToken
	return
}
