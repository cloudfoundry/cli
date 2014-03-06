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
	Authenticate(credentials map[string]string) (apiErr errors.Error)
	RefreshAuthToken() (updatedToken string, apiErr errors.Error)
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

func (uaa UAAAuthenticationRepository) Authenticate(credentials map[string]string) (apiErr errors.Error) {
	data := url.Values{
		"grant_type": {"password"},
		"scope":      {""},
	}
	for key, val := range credentials {
		data[key] = []string{val}
	}

	apiErr = uaa.getAuthToken(data)
	switch response := apiErr.(type) {
	case errors.HttpError:
		if response.StatusCode() == 401 {
			apiErr = errors.NewErrorWithMessage("Password is incorrect, please try again.")
		}
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

func (uaa UAAAuthenticationRepository) GetLoginPrompts() (prompts map[string]configuration.AuthPrompt, apiErr errors.Error) {
	url := fmt.Sprintf("%s/login", uaa.config.AuthorizationEndpoint())
	resource := &LoginResource{}
	apiErr = uaa.gateway.GetResource(url, "", resource)
	prompts = resource.ToModel()
	return
}

func (uaa UAAAuthenticationRepository) RefreshAuthToken() (updatedToken string, apiErr errors.Error) {
	data := url.Values{
		"refresh_token": {uaa.config.RefreshToken()},
		"grant_type":    {"refresh_token"},
		"scope":         {""},
	}

	apiErr = uaa.getAuthToken(data)
	updatedToken = uaa.config.AccessToken()

	if apiErr != nil {
		fmt.Printf("%s\n\n", terminal.NotLoggedInText())
		os.Exit(1)
	}

	return
}

func (uaa UAAAuthenticationRepository) getAuthToken(data url.Values) errors.Error {
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
	request, err := uaa.gateway.NewRequest("POST", path, "Basic "+base64.StdEncoding.EncodeToString([]byte("cf:")), strings.NewReader(data.Encode()))
	if err != nil {
		return errors.NewErrorWithError("Failed to start oauth request", err)
	}
	request.HttpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response := new(AuthenticationResponse)
	_, err = uaa.gateway.PerformRequestForJSONResponse(request, &response)

	switch err.(type) {
	case nil:
	case errors.HttpError:
		return err
	default:
		return errors.NewErrorWithError("auth request failed", err)
	}

	if response.Error.Code != "" {
		return errors.NewError("Authentication Server error: "+response.Error.Description, response.Error.Code)
	}

	uaa.config.SetAccessToken(fmt.Sprintf("%s %s", response.TokenType, response.AccessToken))
	uaa.config.SetRefreshToken(response.RefreshToken)

	return nil
}
