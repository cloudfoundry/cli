package testhelpers

import (
	"cf"
	"cf/net"
)

type FakeAuthTokenRepo struct {
	CreatedServiceAuthToken cf.ServiceAuthToken

	FindAllAuthTokens []cf.ServiceAuthToken
}

func (repo *FakeAuthTokenRepo) Create(authToken cf.ServiceAuthToken) (apiResponse net.ApiResponse) {
	repo.CreatedServiceAuthToken = authToken
	return
}

func (repo *FakeAuthTokenRepo) FindAll() (authTokens []cf.ServiceAuthToken, apiResponse net.ApiResponse) {
	authTokens = repo.FindAllAuthTokens
	return
}
