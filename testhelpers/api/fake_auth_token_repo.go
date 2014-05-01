package api

import "github.com/cloudfoundry/cli/cf/models"

type FakeAuthTokenRepo struct {
	CreatedServiceAuthTokenFields models.ServiceAuthTokenFields

	FindAllAuthTokens []models.ServiceAuthTokenFields

	FindByLabelAndProviderLabel                  string
	FindByLabelAndProviderProvider               string
	FindByLabelAndProviderServiceAuthTokenFields models.ServiceAuthTokenFields
	FindByLabelAndProviderApiResponse            error

	UpdatedServiceAuthTokenFields models.ServiceAuthTokenFields

	DeletedServiceAuthTokenFields models.ServiceAuthTokenFields
}

func (repo *FakeAuthTokenRepo) Create(authToken models.ServiceAuthTokenFields) (apiErr error) {
	repo.CreatedServiceAuthTokenFields = authToken
	return
}

func (repo *FakeAuthTokenRepo) FindAll() (authTokens []models.ServiceAuthTokenFields, apiErr error) {
	authTokens = repo.FindAllAuthTokens
	return
}
func (repo *FakeAuthTokenRepo) FindByLabelAndProvider(label, provider string) (authToken models.ServiceAuthTokenFields, apiErr error) {
	repo.FindByLabelAndProviderLabel = label
	repo.FindByLabelAndProviderProvider = provider

	authToken = repo.FindByLabelAndProviderServiceAuthTokenFields
	apiErr = repo.FindByLabelAndProviderApiResponse
	return
}

func (repo *FakeAuthTokenRepo) Delete(authToken models.ServiceAuthTokenFields) (apiErr error) {
	repo.DeletedServiceAuthTokenFields = authToken
	return
}

func (repo *FakeAuthTokenRepo) Update(authToken models.ServiceAuthTokenFields) (apiErr error) {
	repo.UpdatedServiceAuthTokenFields = authToken
	return
}
