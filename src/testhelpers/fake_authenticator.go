package testhelpers

import (
	"cf/configuration"
	"cf/api"
)

type FakeAuthenticator struct {
	ConfigRepo FakeConfigRepository

	Config *configuration.Configuration
	Email string
	Password string

	AuthError bool
	AccessToken string
	RefreshToken string
}

func (auth *FakeAuthenticator) Authenticate(email string, password string) (apiErr *api.ApiError) {
	auth.Config, _ = auth.ConfigRepo.Get()
	auth.Email = email
	auth.Password = password

	if auth.AccessToken == "" {
		auth.AccessToken = "BEARER some_access_token"
	}

	auth.Config.AccessToken = auth.AccessToken
	auth.Config.RefreshToken = auth.RefreshToken
	auth.ConfigRepo.Save()

	if auth.AuthError {
		apiErr =  &api.ApiError{Message: "Error authenticating."}
	}
	return
}

func (auth *FakeAuthenticator) RefreshAuthToken() (updatedToken string, apiErr *api.ApiError) {
	return
}
