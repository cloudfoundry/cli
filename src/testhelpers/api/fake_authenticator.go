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
		ApiResponse errors.Error
		Prompts     map[string]configuration.AuthPrompt
	}

	AuthError    bool
	AccessToken  string
	RefreshToken string
}

func (auth *FakeAuthenticationRepository) Authenticate(credentials map[string]string) (apiResponse errors.Error) {
	auth.AuthenticateArgs.Credentials = credentials

	if auth.AuthError {
		apiResponse = errors.NewErrorWithMessage("Error authenticating.")
		return
	}

	if auth.AccessToken == "" {
		auth.AccessToken = "BEARER some_access_token"
	}

	auth.Config.SetAccessToken(auth.AccessToken)
	auth.Config.SetRefreshToken(auth.RefreshToken)

	return
}

func (auth *FakeAuthenticationRepository) RefreshAuthToken() (updatedToken string, apiResponse errors.HttpError) {
	return
}

func (auth *FakeAuthenticationRepository) GetLoginPrompts() (prompts map[string]configuration.AuthPrompt, apiResponse errors.Error) {
	prompts = auth.GetLoginPromptsReturns.Prompts
	apiResponse = auth.GetLoginPromptsReturns.ApiResponse
	return
}
