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

type AuthenticationRepository interface {
	Authenticate(email string, password string) (apiResponse net.ApiResponse)
	RefreshAuthToken() (updatedToken string, apiResponse net.ApiResponse)
}

type UAAAuthenticationRepository struct {
	configRepo configuration.ConfigurationRepository
	config     *configuration.Configuration
	gateway    net.Gateway
}

func NewUAAAuthenticationRepository(gateway net.Gateway, configRepo configuration.ConfigurationRepository) (uaa UAAAuthenticationRepository) {
	uaa.gateway = gateway
	uaa.configRepo = configRepo
	uaa.config, _ = configRepo.Get()
	return
}

func (uaa UAAAuthenticationRepository) Authenticate(email string, password string) (apiResponse net.ApiResponse) {
	data := url.Values{
		"username":   {email},
		"password":   {password},
		"grant_type": {"password"},
		"scope":      {""},
	}

	apiResponse = uaa.getAuthToken(data)
	if apiResponse.IsNotSuccessful() && apiResponse.StatusCode == 401 {
		apiResponse.Message = "Password is incorrect, please try again."
	}
	return
}

func (uaa UAAAuthenticationRepository) RefreshAuthToken() (updatedToken string, apiResponse net.ApiResponse) {
	data := url.Values{
		"refresh_token": {uaa.config.RefreshToken},
		"grant_type":    {"refresh_token"},
		"scope":         {""},
	}

	apiResponse = uaa.getAuthToken(data)
	updatedToken = uaa.config.AccessToken

	if apiResponse.IsError() {
		fmt.Printf("%s\n\n", terminal.NotLoggedInText())
		os.Exit(1)
	}

	return
}

func (uaa UAAAuthenticationRepository) getAuthToken(data url.Values) (apiResponse net.ApiResponse) {
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
	request, apiResponse := uaa.gateway.NewRequest("POST", path, "Basic "+base64.StdEncoding.EncodeToString([]byte("cf:")), strings.NewReader(data.Encode()))
	if apiResponse.IsNotSuccessful() {
		return
	}
	request.HttpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response := new(AuthenticationResponse)
	_, apiResponse = uaa.gateway.PerformRequestForJSONResponse(request, &response)

	if apiResponse.IsNotSuccessful() {
		return
	}

	if response.Error.Code != "" {
		apiResponse = net.NewApiResponseWithMessage("Authentication Server error: %s", response.Error.Description)
		return
	}

	uaa.config.AccessToken = fmt.Sprintf("%s %s", response.TokenType, response.AccessToken)
	uaa.config.RefreshToken = response.RefreshToken
	err := uaa.configRepo.Save()
	if err != nil {
		apiResponse = net.NewApiResponseWithError("Error setting configuration", err)
	}

	return
}
