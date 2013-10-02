package testhelpers

import (
	"cf/configuration"
	"cf/net"
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

func (auth *FakeAuthenticator) Authenticate(email string, password string) (apiStatus net.ApiStatus) {
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
		apiStatus =  net.NewApiStatusWithMessage("Error authenticating.")
	}
	return
}

func (auth *FakeAuthenticator) RefreshAuthToken() (updatedToken string, apiStatus net.ApiStatus) {
	return
}
