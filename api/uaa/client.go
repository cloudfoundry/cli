package uaa

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

//go:generate counterfeiter . AuthenticationStore

type AuthenticationStore interface {
	ClientID() string
	ClientSecret() string
	SkipSSLValidation() bool

	AccessToken() string
	RefreshToken() string
	SetAccessToken(token string)
}

type Client struct {
	store AuthenticationStore
	URL   string
}

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

func (refreshTokenResponse RefreshTokenResponse) AuthorizationToken() string {
	return fmt.Sprintf("%s %s", refreshTokenResponse.TokenType, refreshTokenResponse.AccessToken)
}

func (client *Client) RefreshToken() error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: client.store.SkipSSLValidation(),
		},
	}

	httpClient := http.Client{Transport: tr}

	urlValues := url.Values{
		"client_id":     {client.store.ClientID()},
		"client_secret": {client.store.ClientSecret()},
		"grant_type":    {"refresh_token"},
		"refresh_token": {client.store.RefreshToken()},
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/oauth/token", client.URL), strings.NewReader(urlValues.Encode()))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		return err
	}

	uaaResponse, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	var refreshResponse RefreshTokenResponse
	rawBytes, err := ioutil.ReadAll(uaaResponse.Body)
	if err != nil {
		return err
	}
	json.Unmarshal(rawBytes, &refreshResponse)
	client.store.SetAccessToken(refreshResponse.AuthorizationToken())
	return nil
}

func (client *Client) AccessToken() string {
	return client.store.AccessToken()
}

func NewClient(URL string, store AuthenticationStore) *Client {
	return &Client{
		store: store,
		URL:   URL,
	}
}
