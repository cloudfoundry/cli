package api

import (
	"cf/configuration"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
)

type Authenticator interface {
	Authenticate(email string, password string) (apiErr *ApiError)
	RefreshAuthToken() (updatedToken string, apiErr *ApiError)
}

type UAAAuthenticator struct {
	configRepo configuration.ConfigurationRepository
	config     *configuration.Configuration
}

func NewUAAAuthenticator(configRepo configuration.ConfigurationRepository) (uaa UAAAuthenticator) {
	uaa.configRepo = configRepo
	uaa.config, _ = configRepo.Get()
	return
}

func (uaa UAAAuthenticator) Authenticate(email string, password string) (apiErr *ApiError) {
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

func (uaa UAAAuthenticator) RefreshAuthToken() (updatedToken string, apiErr *ApiError) {
	data := url.Values{
		"refresh_token": {uaa.config.RefreshToken},
		"grant_type":    {"refresh_token"},
		"scope":         {""},
	}

	apiErr = uaa.getAuthToken(data)
	updatedToken = uaa.config.AccessToken

	if apiErr != nil && apiErr.StatusCode == 401 {
		apiErr.Message = "Session expired, please login."
	}

	return
}

func (uaa UAAAuthenticator) getAuthToken(data url.Values) (apiErr *ApiError) {
	type AuthenticationResponse struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		RefreshToken string `json:"refresh_token"`
	}

	path := fmt.Sprintf("%s/oauth/token", uaa.config.AuthorizationEndpoint)
	request, apiErr := NewRequest("POST", path, "Basic "+base64.StdEncoding.EncodeToString([]byte("cf:")), strings.NewReader(data.Encode()))
	if apiErr != nil {
		return
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response := new(AuthenticationResponse)
	apiErr = PerformRequestAndParseResponse(request, &response)

	if apiErr != nil {
		return
	}

	uaa.config.AccessToken = fmt.Sprintf("%s %s", response.TokenType, response.AccessToken)
	uaa.config.RefreshToken = response.RefreshToken
	err := uaa.configRepo.Save()
	if err != nil {
		apiErr = NewApiErrorWithError("Error setting configuration", err)
	}

	return
}
