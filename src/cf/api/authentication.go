package api

import (
	"cf/configuration"
	"cf/errors"
	"cf/net"
	"cf/terminal"
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"strings"
)

type AuthenticationRepository interface {
	Authenticate(credentials map[string]string) (apiResponse errors.Error)
	RefreshAuthToken() (updatedToken string, apiResponse errors.HttpError)
	GetLoginPrompts() (map[string]configuration.AuthPrompt, errors.Error)
}

type UAAAuthenticationRepository struct {
	config  configuration.ReadWriter
	gateway net.Gateway
}

func NewUAAAuthenticationRepository(gateway net.Gateway, config configuration.ReadWriter) (uaa UAAAuthenticationRepository) {
	uaa.gateway = gateway
	uaa.config = config
	return
}

func (uaa UAAAuthenticationRepository) Authenticate(credentials map[string]string) (apiResponse errors.Error) {
	data := url.Values{
		"grant_type": {"password"},
		"scope":      {""},
	}
	for key, val := range credentials {
		data[key] = []string{val}
	}

	apiResponse = uaa.getAuthToken(data)
	if apiResponse != nil && apiResponse.StatusCode() == 401 {
		apiResponse = errors.NewErrorWithMessage("Password is incorrect, please try again.")
	}
	return
}

type LoginResource struct {
	Prompts map[string][]string
}

var knownAuthPromptTypes = map[string]configuration.AuthPromptType{
	"text":     configuration.AuthPromptTypeText,
	"password": configuration.AuthPromptTypePassword,
}

func (r *LoginResource) ToModel() (prompts map[string]configuration.AuthPrompt) {
	prompts = make(map[string]configuration.AuthPrompt)
	for key, val := range r.Prompts {
		prompts[key] = configuration.AuthPrompt{
			Type:        knownAuthPromptTypes[val[0]],
			DisplayName: val[1],
		}
	}
	return
}

func (uaa UAAAuthenticationRepository) GetLoginPrompts() (prompts map[string]configuration.AuthPrompt, apiResponse errors.Error) {
	url := fmt.Sprintf("%s/login", uaa.config.AuthorizationEndpoint())
	resource := &LoginResource{}
	apiResponse = uaa.gateway.GetResource(url, "", resource)
	prompts = resource.ToModel()
	return
}

func (uaa UAAAuthenticationRepository) RefreshAuthToken() (updatedToken string, apiResponse errors.HttpError) {
	data := url.Values{
		"refresh_token": {uaa.config.RefreshToken()},
		"grant_type":    {"refresh_token"},
		"scope":         {""},
	}

	apiResponse = uaa.getAuthToken(data)
	updatedToken = uaa.config.AccessToken()

	if apiResponse != nil {
		fmt.Printf("%s\n\n", terminal.NotLoggedInText())
		os.Exit(1)
	}

	return
}

func (uaa UAAAuthenticationRepository) getAuthToken(data url.Values) (httpError errors.HttpError) {
	type uaaErrorResponse struct {
		Code        string `json:"error"`
		Description string `json:"error_description"`
	}

	type AuthenticationResponse struct {
		AccessToken  string           `json:"access_token"`
		TokenType    string           `json:"token_type"`
		RefreshToken string           `json:"refresh_token"`
		Error        uaaErrorResponse `json:"error"`
	}

	path := fmt.Sprintf("%s/oauth/token", uaa.config.AuthorizationEndpoint())
	request, apiResponse := uaa.gateway.NewRequest("POST", path, "Basic "+base64.StdEncoding.EncodeToString([]byte("cf:")), strings.NewReader(data.Encode()))
	if apiResponse != nil {
		httpError = errors.NewHTTPErrorWithError("Failed to start oauth request", apiResponse)
		return
	}
	request.HttpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response := new(AuthenticationResponse)
	_, apiResponse = uaa.gateway.PerformRequestForJSONResponse(request, &response)

	if apiResponse != nil {
		httpError = errors.NewHTTPErrorWithError("auth request failed", apiResponse)
		return
	}

	if response.Error.Code != "" {
		apiResponse = errors.NewErrorWithMessage("Authentication Server error: %s", response.Error.Description)
		return
	}

	uaa.config.SetAccessToken(fmt.Sprintf("%s %s", response.TokenType, response.AccessToken))
	uaa.config.SetRefreshToken(response.RefreshToken)

	return
}
