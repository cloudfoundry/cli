package api

import (
	"cf/configuration"
	"cf/errors"
)

type FakeAuthenticationRepository struct {
	Config           configuration.ReadWriter
	AuthenticateArgs struct {
		Credentials map[string]string
	}
	GetLoginPromptsReturns struct {
		Error   errors.Error
		Prompts map[string]configuration.AuthPrompt
	}

	AuthError    bool
	AccessToken  string
	RefreshToken string
}

func (auth *FakeAuthenticationRepository) Authenticate(credentials map[string]string) (apiErr errors.Error) {
	auth.AuthenticateArgs.Credentials = credentials

	if auth.AuthError {
		apiErr = errors.NewErrorWithMessage("Error authenticating.")
		return
	}

	if auth.AccessToken == "" {
		auth.AccessToken = "BEARER some_access_token"
	}

	auth.Config.SetAccessToken(auth.AccessToken)
	auth.Config.SetRefreshToken(auth.RefreshToken)

	return
}

func (auth *FakeAuthenticationRepository) RefreshAuthToken() (updatedToken string, apiErr errors.Error) {
	return
}

func (auth *FakeAuthenticationRepository) GetLoginPromptsAndSaveUAAServerURL() (prompts map[string]configuration.AuthPrompt, apiErr errors.Error) {
	prompts = auth.GetLoginPromptsReturns.Prompts
	apiErr = auth.GetLoginPromptsReturns.Error
	return
}
