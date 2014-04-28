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

type TokenRefresher interface {
	RefreshAuthToken() (updatedToken string, apiErr error)
}

type AuthenticationRepository interface {
	RefreshAuthToken() (updatedToken string, apiErr error)
	Authenticate(credentials map[string]string) (apiErr error)
	GetLoginPromptsAndSaveUAAServerURL() (map[string]configuration.AuthPrompt, error)
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

func (uaa UAAAuthenticationRepository) Authenticate(credentials map[string]string) (apiErr error) {
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
			apiErr = errors.New("Credentials were rejected, please try again.")
		}
	}

	return
}

type LoginResource struct {
	Prompts map[string][]string
	Links   map[string]string
}

var knownAuthPromptTypes = map[string]configuration.AuthPromptType{
	"text":     configuration.AuthPromptTypeText,
	"password": configuration.AuthPromptTypePassword,
}

func (r *LoginResource) parsePrompts() (prompts map[string]configuration.AuthPrompt) {
	prompts = make(map[string]configuration.AuthPrompt)
	for key, val := range r.Prompts {
		prompts[key] = configuration.AuthPrompt{
			Type:        knownAuthPromptTypes[val[0]],
			DisplayName: val[1],
		}
	}
	return
}

func (uaa UAAAuthenticationRepository) GetLoginPromptsAndSaveUAAServerURL() (prompts map[string]configuration.AuthPrompt, apiErr error) {
	url := fmt.Sprintf("%s/login", uaa.config.AuthenticationEndpoint())
	resource := &LoginResource{}
	apiErr = uaa.gateway.GetResource(url, resource)

	prompts = resource.parsePrompts()
	if resource.Links["uaa"] == "" {
		uaa.config.SetUaaEndpoint(uaa.config.AuthenticationEndpoint())
	} else {
		uaa.config.SetUaaEndpoint(resource.Links["uaa"])
	}
	return
}

func (uaa UAAAuthenticationRepository) RefreshAuthToken() (updatedToken string, apiErr error) {
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

func (uaa UAAAuthenticationRepository) getAuthToken(data url.Values) error {
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

	path := fmt.Sprintf("%s/oauth/token", uaa.config.AuthenticationEndpoint())
	request, err := uaa.gateway.NewRequest("POST", path, "Basic "+base64.StdEncoding.EncodeToString([]byte("cf:")), strings.NewReader(data.Encode()))
	if err != nil {
		return errors.NewWithError("Failed to start oauth request", err)
	}
	request.HttpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response := new(AuthenticationResponse)
	_, err = uaa.gateway.PerformRequestForJSONResponse(request, &response)

	switch err.(type) {
	case nil:
	case errors.HttpError:
		return err
	default:
		return errors.NewWithError("auth request failed", err)
	}

	// TODO: get the actual status code
	if response.Error.Code != "" {
		return errors.NewHttpError(0, response.Error.Code, response.Error.Description)
	}

	uaa.config.SetAccessToken(fmt.Sprintf("%s %s", response.TokenType, response.AccessToken))
	uaa.config.SetRefreshToken(response.RefreshToken)

	return nil
}
