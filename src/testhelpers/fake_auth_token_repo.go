package testhelpers

import (
	"cf"
	"cf/net"
)

type FakeAuthTokenRepo struct {
	CreatedServiceAuthToken cf.ServiceAuthToken
	
	FindAllAuthTokens []cf.ServiceAuthToken

	FindByNameServiceAuthToken cf.ServiceAuthToken
	FindByNameTokenNameKey cf.ServiceAuthTokenNameKey

	UpdatedServiceAuthToken cf.ServiceAuthToken

	DeletedServiceAuthToken cf.ServiceAuthToken
}

func (repo *FakeAuthTokenRepo) Create(authToken cf.ServiceAuthToken) (apiResponse net.ApiResponse) {
	repo.CreatedServiceAuthToken = authToken
	return
}

func (repo *FakeAuthTokenRepo) FindAll() (authTokens []cf.ServiceAuthToken, apiResponse net.ApiResponse) {
	authTokens = repo.FindAllAuthTokens
	return
}
func (repo *FakeAuthTokenRepo) FindByName(tokenName cf.ServiceAuthTokenNameKey) (authToken cf.ServiceAuthToken, apiResponse net.ApiResponse) {
	repo.FindByNameTokenNameKey = tokenName
	authToken = repo.FindByNameServiceAuthToken
	return
}

func (repo  *FakeAuthTokenRepo) Delete(authToken cf.ServiceAuthToken) (apiResponse net.ApiResponse) {
	repo.DeletedServiceAuthToken = authToken
	return
}

func (repo *FakeAuthTokenRepo) Update(authToken cf.ServiceAuthToken) (apiResponse net.ApiResponse) {
	repo.UpdatedServiceAuthToken = authToken
	return
}
