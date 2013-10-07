package testhelpers

import (
	"cf/configuration"
	"cf/net"
)

type FakeAuthenticationRepository struct {
	ConfigRepo FakeConfigRepository

	Config *configuration.Configuration
	Email string
	Password string

	AuthError bool
	AccessToken string
	RefreshToken string
}

func (auth *FakeAuthenticationRepository) Authenticate(email string, password string) (apiResponse net.ApiResponse) {
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
		apiResponse =  net.NewApiStatusWithMessage("Error authenticating.")
	}
	return
}

func (auth *FakeAuthenticationRepository) RefreshAuthToken() (updatedToken string, apiResponse net.ApiResponse) {
	return
}
