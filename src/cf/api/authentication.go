package api

import (
	"cf/configuration"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Authenticator interface {
	Authenticate(config *configuration.Configuration, email string, password string) (err error)
}

type UAAAuthenticator struct {
}

func (uaa UAAAuthenticator) Authenticate(config *configuration.Configuration, email string, password string) (err error) {
	type AuthenticationResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}

	data := url.Values{
		"username":   {email},
		"password":   {password},
		"grant_type": {"password"},
		"scope":      {""},
	}

	path := fmt.Sprintf("%s/oauth/token", config.AuthorizationEndpoint)
	request, err := http.NewRequest("POST", path, strings.NewReader(data.Encode()))
	if err != nil {
		return
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("cf:")))

	response := new(AuthenticationResponse)
	err = PerformRequestForBody(request, &response)

	if err != nil {
		return
	}

	config.AccessToken = fmt.Sprintf("%s %s", response.TokenType, response.AccessToken)
	return config.Save()
}
