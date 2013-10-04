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
	Authenticate(email string, password string) (apiStatus net.ApiStatus)
	RefreshAuthToken() (updatedToken string, apiStatus net.ApiStatus)
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

func (uaa UAAAuthenticator) Authenticate(email string, password string) (apiStatus net.ApiStatus) {
	data := url.Values{
		"username":   {email},
		"password":   {password},
		"grant_type": {"password"},
		"scope":      {""},
	}

	apiStatus = uaa.getAuthToken(data)
	if apiStatus.NotSuccessful() && apiStatus.StatusCode == 401 {
		apiStatus.Message = "Password is incorrect, please try again."
	}
	return
}

func (uaa UAAAuthenticator) RefreshAuthToken() (updatedToken string, apiStatus net.ApiStatus) {
	data := url.Values{
		"refresh_token": {uaa.config.RefreshToken},
		"grant_type":    {"refresh_token"},
		"scope":         {""},
	}

	apiStatus = uaa.getAuthToken(data)
	updatedToken = uaa.config.AccessToken

	if apiStatus.IsError() {
		fmt.Printf("%s\n\n", terminal.NotLoggedInText())
		os.Exit(1)
	}

	return
}

func (uaa UAAAuthenticator) getAuthToken(data url.Values) (apiStatus net.ApiStatus) {
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

	path := fmt.Sprintf("%s/oauth/token", uaa.config.AuthorizationEndpoint)
	request, apiStatus := uaa.gateway.NewRequest("POST", path, "Basic "+base64.StdEncoding.EncodeToString([]byte("cf:")), strings.NewReader(data.Encode()))
	if apiStatus.NotSuccessful() {
		return
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response := new(AuthenticationResponse)
	_, apiStatus = uaa.gateway.PerformRequestForJSONResponse(request, &response)

	if apiStatus.NotSuccessful() {
		return
	}

	if response.Error.Code != "" {
		apiStatus = net.NewApiStatusWithMessage("Authentication Server error: %s", response.Error.Description)
		return
	}

	uaa.config.AccessToken = fmt.Sprintf("%s %s", response.TokenType, response.AccessToken)
	uaa.config.RefreshToken = response.RefreshToken
	err := uaa.configRepo.Save()
	if err != nil {
		apiStatus = net.NewApiStatusWithError("Error setting configuration", err)
	}

	return
}
