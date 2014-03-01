package api

import (
	"cf/errors"
	"cf/models"
)

type FakeAuthTokenRepo struct {
	CreatedServiceAuthTokenFields models.ServiceAuthTokenFields

	FindAllAuthTokens []models.ServiceAuthTokenFields

	FindByLabelAndProviderLabel                  string
	FindByLabelAndProviderProvider               string
	FindByLabelAndProviderServiceAuthTokenFields models.ServiceAuthTokenFields
	FindByLabelAndProviderApiResponse            errors.Error

	UpdatedServiceAuthTokenFields models.ServiceAuthTokenFields

	DeletedServiceAuthTokenFields models.ServiceAuthTokenFields
}

func (repo *FakeAuthTokenRepo) Create(authToken models.ServiceAuthTokenFields) (apiErr errors.Error) {
	repo.CreatedServiceAuthTokenFields = authToken
	return
}

func (repo *FakeAuthTokenRepo) FindAll() (authTokens []models.ServiceAuthTokenFields, apiErr errors.Error) {
	authTokens = repo.FindAllAuthTokens
	return
}
func (repo *FakeAuthTokenRepo) FindByLabelAndProvider(label, provider string) (authToken models.ServiceAuthTokenFields, apiErr errors.Error) {
	repo.FindByLabelAndProviderLabel = label
	repo.FindByLabelAndProviderProvider = provider

	authToken = repo.FindByLabelAndProviderServiceAuthTokenFields
	apiErr = repo.FindByLabelAndProviderApiResponse
	return
}

func (repo *FakeAuthTokenRepo) Delete(authToken models.ServiceAuthTokenFields) (apiErr errors.Error) {
	repo.DeletedServiceAuthTokenFields = authToken
	return
}

func (repo *FakeAuthTokenRepo) Update(authToken models.ServiceAuthTokenFields) (apiErr errors.Error) {
	repo.UpdatedServiceAuthTokenFields = authToken
	return
}
