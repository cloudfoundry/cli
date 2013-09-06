package api

import (
	"cf/configuration"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
)

type Authenticator interface {
	Authenticate(email string, password string) (err error)
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

func (uaa UAAAuthenticator) Authenticate(email string, password string) (err error) {
	type AuthenticationResponse struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		RefreshToken string `json:"refresh_token"`
	}

	data := url.Values{
		"username":   {email},
		"password":   {password},
		"grant_type": {"password"},
		"scope":      {""},
	}

	path := fmt.Sprintf("%s/oauth/token", uaa.config.AuthorizationEndpoint)
	request, err := NewRequest("POST", path, "Basic "+base64.StdEncoding.EncodeToString([]byte("cf:")), strings.NewReader(data.Encode()))
	if err != nil {
		return
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response := new(AuthenticationResponse)
	_, err = PerformRequestAndParseResponse(request, &response)

	if err != nil {
		return
	}

	uaa.config.AccessToken = fmt.Sprintf("%s %s", response.TokenType, response.AccessToken)
	uaa.config.RefreshToken = response.RefreshToken
	return uaa.configRepo.Save()
}
