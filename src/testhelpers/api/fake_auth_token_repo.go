package api

import (
	"cf"
	"cf/net"
)

type FakeAuthTokenRepo struct {
	CreatedServiceAuthToken cf.ServiceAuthToken
	
	FindAllAuthTokens []cf.ServiceAuthToken

	FindByLabelAndProviderLabel string
	FindByLabelAndProviderProvider string
	FindByLabelAndProviderServiceAuthToken cf.ServiceAuthToken
	FindByLabelAndProviderApiResponse net.ApiResponse

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
func (repo *FakeAuthTokenRepo) FindByLabelAndProvider(label, provider string) (authToken cf.ServiceAuthToken, apiResponse net.ApiResponse) {
	repo.FindByLabelAndProviderLabel = label
	repo.FindByLabelAndProviderProvider = provider

	authToken = repo.FindByLabelAndProviderServiceAuthToken
	apiResponse = repo.FindByLabelAndProviderApiResponse
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
