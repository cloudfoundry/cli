package api

import (
	"cf/configuration"
	"cf/net"
	"cf/terminal"
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"strings"
)

type Authenticator interface {
	Authenticate(email string, password string) (apiErr *net.ApiError)
	RefreshAuthToken() (updatedToken string, apiErr *net.ApiError)
}

type UAAAuthenticator struct {
	configRepo configuration.ConfigurationRepository
	config     *configuration.Configuration
	gateway    net.Gateway
}

func NewUAAAuthenticator(gateway net.Gateway, configRepo configuration.ConfigurationRepository) (uaa UAAAuthenticator) {
	uaa.gateway = gateway
	uaa.configRepo = configRepo
	uaa.config, _ = configRepo.Get()
	return
}

func (uaa UAAAuthenticator) Authenticate(email string, password string) (apiErr *net.ApiError) {
	data := url.Values{
		"username":   {email},
		"password":   {password},
		"grant_type": {"password"},
		"scope":      {""},
	}

	apiErr = uaa.getAuthToken(data)
	if apiErr != nil && apiErr.StatusCode == 401 {
		apiErr.Message = "Password is incorrect, please try again."
	}
	return
}

func (uaa UAAAuthenticator) RefreshAuthToken() (updatedToken string, apiErr *net.ApiError) {
	data := url.Values{
		"refresh_token": {uaa.config.RefreshToken},
		"grant_type":    {"refresh_token"},
		"scope":         {""},
	}

	apiErr = uaa.getAuthToken(data)
	updatedToken = uaa.config.AccessToken

	if apiErr != nil && apiErr.StatusCode == 401 {
		fmt.Printf("%s\n\n", terminal.NotLoggedInText())
		os.Exit(1)
	}

	return
}

func (uaa UAAAuthenticator) getAuthToken(data url.Values) (apiErr *net.ApiError) {
	type AuthenticationResponse struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		RefreshToken string `json:"refresh_token"`
	}

	path := fmt.Sprintf("%s/oauth/token", uaa.config.AuthorizationEndpoint)
	request, apiErr := uaa.gateway.NewRequest("POST", path, "Basic "+base64.StdEncoding.EncodeToString([]byte("cf:")), strings.NewReader(data.Encode()))
	if apiErr != nil {
		return
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response := new(AuthenticationResponse)
	apiErr = uaa.gateway.PerformRequestForJSONResponse(request, &response)

	if apiErr != nil {
		return
	}

	uaa.config.AccessToken = fmt.Sprintf("%s %s", response.TokenType, response.AccessToken)
	uaa.config.RefreshToken = response.RefreshToken
	err := uaa.configRepo.Save()
	if err != nil {
		apiErr = net.NewApiErrorWithError("Error setting configuration", err)
	}

	return
}
