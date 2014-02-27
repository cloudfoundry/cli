package api

import (
	"cf/configuration"
	"cf/net"
)

type FakeAuthenticationRepository struct {
	Config           configuration.ReadWriter
	AuthenticateArgs struct {
		Credentials map[string]string
	}
	GetLoginPromptsReturns struct {
		ApiResponse net.ApiResponse
		Prompts     map[string]configuration.AuthPrompt
	}

	AuthError    bool
	AccessToken  string
	RefreshToken string
}

func (auth *FakeAuthenticationRepository) Authenticate(credentials map[string]string) (apiResponse net.ApiResponse) {
	auth.AuthenticateArgs.Credentials = credentials

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

func (auth *FakeAuthenticationRepository) GetLoginPrompts() (prompts map[string]configuration.AuthPrompt, apiResponse net.ApiResponse) {
	prompts = auth.GetLoginPromptsReturns.Prompts
	apiResponse = auth.GetLoginPromptsReturns.ApiResponse
	return
}
