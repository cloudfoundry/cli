package api

import (
	"cf/configuration"
	"cf/net"
)

type FakeAuthenticationRepository struct {
	Config   configuration.ReadWriter
	Email    string
	Password string

	AuthError    bool
	AccessToken  string
	RefreshToken string
}

func (auth *FakeAuthenticationRepository) Authenticate(email string, password string) (apiResponse net.ApiResponse) {
	auth.Email = email
	auth.Password = password

	if auth.AuthError {
		apiResponse = net.NewApiResponseWithMessage("Error authenticating.")
		return
	}

	if auth.AccessToken == "" {
		auth.AccessToken = "BEARER some_access_token"
	}

	auth.Config.SetAccessToken(auth.AccessToken)
	auth.Config.SetRefreshToken(auth.RefreshToken)

	return
}

func (auth *FakeAuthenticationRepository) RefreshAuthToken() (updatedToken string, apiResponse net.ApiResponse) {
	return
}
