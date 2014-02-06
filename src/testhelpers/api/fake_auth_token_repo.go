package api

import (
"cf/models"
	"cf/net"
)

type FakeAuthTokenRepo struct {
	CreatedServiceAuthTokenFields models.ServiceAuthTokenFields

	FindAllAuthTokens []models.ServiceAuthTokenFields

	FindByLabelAndProviderLabel                  string
	FindByLabelAndProviderProvider               string
	FindByLabelAndProviderServiceAuthTokenFields models.ServiceAuthTokenFields
	FindByLabelAndProviderApiResponse            net.ApiResponse

	UpdatedServiceAuthTokenFields models.ServiceAuthTokenFields

	DeletedServiceAuthTokenFields models.ServiceAuthTokenFields
}

func (repo *FakeAuthTokenRepo) Create(authToken models.ServiceAuthTokenFields) (apiResponse net.ApiResponse) {
	repo.CreatedServiceAuthTokenFields = authToken
	return
}

func (repo *FakeAuthTokenRepo) FindAll() (authTokens []models.ServiceAuthTokenFields, apiResponse net.ApiResponse) {
	authTokens = repo.FindAllAuthTokens
	return
}
func (repo *FakeAuthTokenRepo) FindByLabelAndProvider(label, provider string) (authToken models.ServiceAuthTokenFields, apiResponse net.ApiResponse) {
	repo.FindByLabelAndProviderLabel = label
	repo.FindByLabelAndProviderProvider = provider

	authToken = repo.FindByLabelAndProviderServiceAuthTokenFields
	apiResponse = repo.FindByLabelAndProviderApiResponse
	return
}

func (repo *FakeAuthTokenRepo) Delete(authToken models.ServiceAuthTokenFields) (apiResponse net.ApiResponse) {
	repo.DeletedServiceAuthTokenFields = authToken
	return
}

func (repo *FakeAuthTokenRepo) Update(authToken models.ServiceAuthTokenFields) (apiResponse net.ApiResponse) {
	repo.UpdatedServiceAuthTokenFields = authToken
	return
}
