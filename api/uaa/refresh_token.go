package uaa

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"code.cloudfoundry.org/cli/api/uaa/constant"
	"code.cloudfoundry.org/cli/api/uaa/internal"
)

// RefreshedTokens represents the UAA refresh token response.
type RefreshedTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Type         string `json:"token_type"`
}

// AuthorizationToken returns formatted authorization header.
func (refreshTokenResponse RefreshedTokens) AuthorizationToken() string {
	return fmt.Sprintf("%s %s", refreshTokenResponse.Type, refreshTokenResponse.AccessToken)
}

// RefreshAccessToken refreshes the current access token.
func (client *Client) RefreshAccessToken(refreshToken string) (RefreshedTokens, error) {
	var values url.Values

	switch client.config.UAAGrantType() {
	case string(constant.GrantTypeClientCredentials):
		values = client.clientCredentialRefreshBody()
	case "", string(constant.GrantTypePassword): // CLI used to write empty string for grant type in the case of password; preserve compatibility with old config.json files
		values = client.refreshTokenBody(refreshToken)
	}

	body := strings.NewReader(values.Encode())

	request, err := client.newRequest(requestOptions{
		RequestName: internal.PostOAuthTokenRequest,
		Header:      http.Header{"Content-Type": {"application/x-www-form-urlencoded"}},
		Body:        body,
	})
	if err != nil {
		return RefreshedTokens{}, err
	}

	if client.config.UAAGrantType() != string(constant.GrantTypeClientCredentials) {
		request.SetBasicAuth(client.config.UAAOAuthClient(), client.config.UAAOAuthClientSecret())
	}

	var refreshResponse RefreshedTokens
	response := Response{
		Result: &refreshResponse,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return RefreshedTokens{}, err
	}

	return refreshResponse, nil
}

func (client *Client) clientCredentialRefreshBody() url.Values {
	return url.Values{
		"client_id":     {client.config.UAAOAuthClient()},
		"client_secret": {client.config.UAAOAuthClientSecret()},
		"grant_type":    {string(constant.GrantTypeClientCredentials)},
	}
}

func (client *Client) refreshTokenBody(refreshToken string) url.Values {
	return url.Values{
		"refresh_token": {refreshToken},
		"grant_type":    {string(constant.GrantTypeRefreshToken)},
	}
}
