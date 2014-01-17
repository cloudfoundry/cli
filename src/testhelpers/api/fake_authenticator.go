package api

import (
	"cf/configuration"
	"cf/net"
	testconfig "testhelpers/configuration"
)

type FakeAuthenticationRepository struct {
	ConfigRepo testconfig.FakeConfigRepository

	Config   *configuration.Configuration
	Email    string
	Password string

	AuthError    bool
	AccessToken  string
	RefreshToken string
}

func (auth *FakeAuthenticationRepository) Authenticate(email string, password string) (apiResponse net.ApiResponse) {
	auth.Config, _ = auth.ConfigRepo.Get()
	auth.Email = email
	auth.Password = password

	if auth.AuthError {
		apiResponse = net.NewApiResponseWithMessage("Error authenticating.")
		return
	}

	if auth.AccessToken == "" {
		auth.AccessToken = "BEARER some_access_token"
	}

	auth.Config.AccessToken = auth.AccessToken
	auth.Config.RefreshToken = auth.RefreshToken
	auth.ConfigRepo.Save()

	return
}

func (auth *FakeAuthenticationRepository) RefreshAuthToken() (updatedToken string, apiResponse net.ApiResponse) {
	return
}
