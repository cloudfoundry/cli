package fakes

import (
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
)

type FakeAuthenticationRepository struct {
	Config           core_config.ReadWriter
	AuthenticateArgs struct {
		Credentials []map[string]string
	}
	GetLoginPromptsWasCalled bool
	GetLoginPromptsReturns   struct {
		Error   error
		Prompts map[string]core_config.AuthPrompt
	}

	AuthError          bool
	AccessToken        string
	RefreshToken       string
	RefreshTokenCalled bool
	RefreshTokenError  error
}

func (auth *FakeAuthenticationRepository) Authenticate(credentials map[string]string) (apiErr error) {
	auth.AuthenticateArgs.Credentials = append(auth.AuthenticateArgs.Credentials, copyMap(credentials))

	if auth.AuthError {
		apiErr = errors.New("Error authenticating.")
		return
	}

	if auth.AccessToken == "" {
		auth.AccessToken = "BEARER some_access_token"
	}

	auth.Config.SetAccessToken(auth.AccessToken)
	auth.Config.SetRefreshToken(auth.RefreshToken)

	return
}

func (auth *FakeAuthenticationRepository) RefreshAuthToken() (string, error) {
	auth.RefreshTokenCalled = true
	return "", auth.RefreshTokenError
}

func (auth *FakeAuthenticationRepository) GetLoginPromptsAndSaveUAAServerURL() (prompts map[string]core_config.AuthPrompt, apiErr error) {
	auth.GetLoginPromptsWasCalled = true
	prompts = auth.GetLoginPromptsReturns.Prompts
	apiErr = auth.GetLoginPromptsReturns.Error
	return
}

func copyMap(input map[string]string) map[string]string {
	output := map[string]string{}
	for key, val := range input {
		output[key] = val
	}
	return output
}
