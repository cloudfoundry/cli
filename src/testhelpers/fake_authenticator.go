package testhelpers

import (
	"cf/configuration"
	"errors"
)

type FakeAuthenticator struct {
	Config *configuration.Configuration
	Email string
	Password string

	AuthError bool
	AccessToken string
}

func (auth *FakeAuthenticator) Authenticate(config *configuration.Configuration, email string, password string) (err error) {
	auth.Config = config
	auth.Email = email
	auth.Password = password

	if auth.AccessToken == "" {
		auth.AccessToken = "BEARER some_access_token"
	}

	config.AccessToken = auth.AccessToken
	configuration.Save()

	if auth.AuthError {
		err = errors.New("Error authenticating.")
	}

	return
}
