package api

import (
	"cf/configuration"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

type Authenticator interface {
	Authenticate(email string, password string) (err error)
	RefreshAuthToken() (updatedToken string, err error)
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
	data := url.Values{
		"username":   {email},
		"password":   {password},
		"grant_type": {"password"},
		"scope":      {""},
	}

	return uaa.getAuthToken(data)
}

func (uaa UAAAuthenticator) RefreshAuthToken() (updatedToken string, err error) {
	data := url.Values{
		"refresh_token": {uaa.config.RefreshToken},
		"grant_type":    {"refresh_token"},
		"scope":         {""},
	}

	err = uaa.getAuthToken(data)
	updatedToken = uaa.config.AccessToken

	return
}

func (uaa UAAAuthenticator) getAuthToken(data url.Values) (err error) {
	type AuthenticationResponse struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		RefreshToken string `json:"refresh_token"`
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
		if strings.Contains(err.Error(), "status code: 401") {
			err = errors.New("Password in incorrect, please try again.")
		}
		return
	}

	uaa.config.AccessToken = fmt.Sprintf("%s %s", response.TokenType, response.AccessToken)
	uaa.config.RefreshToken = response.RefreshToken
	return uaa.configRepo.Save()
}
