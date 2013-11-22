package api

import (
	"cf"
	"cf/net"
)

type FakeAuthTokenRepo struct {
	CreatedServiceAuthTokenFields cf.ServiceAuthTokenFields
	
	FindAllAuthTokens []cf.ServiceAuthTokenFields

	FindByLabelAndProviderLabel string
	FindByLabelAndProviderProvider string
	FindByLabelAndProviderServiceAuthTokenFields cf.ServiceAuthTokenFields
	FindByLabelAndProviderApiResponse net.ApiResponse

	UpdatedServiceAuthTokenFields cf.ServiceAuthTokenFields

	DeletedServiceAuthTokenFields cf.ServiceAuthTokenFields
}

func (repo *FakeAuthTokenRepo) Create(authToken cf.ServiceAuthTokenFields) (apiResponse net.ApiResponse) {
	repo.CreatedServiceAuthTokenFields = authToken
	return
}

func (repo *FakeAuthTokenRepo) FindAll() (authTokens []cf.ServiceAuthTokenFields, apiResponse net.ApiResponse) {
	authTokens = repo.FindAllAuthTokens
	return
}
func (repo *FakeAuthTokenRepo) FindByLabelAndProvider(label, provider string) (authToken cf.ServiceAuthTokenFields, apiResponse net.ApiResponse) {
	repo.FindByLabelAndProviderLabel = label
	repo.FindByLabelAndProviderProvider = provider

	authToken = repo.FindByLabelAndProviderServiceAuthTokenFields
	apiResponse = repo.FindByLabelAndProviderApiResponse
	return
}

func (repo  *FakeAuthTokenRepo) Delete(authToken cf.ServiceAuthTokenFields) (apiResponse net.ApiResponse) {
	repo.DeletedServiceAuthTokenFields = authToken
	return
}

func (repo *FakeAuthTokenRepo) Update(authToken cf.ServiceAuthTokenFields) (apiResponse net.ApiResponse) {
	repo.UpdatedServiceAuthTokenFields = authToken
	return
}
