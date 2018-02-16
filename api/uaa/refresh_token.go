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
	values := url.Values{
		"client_id":     {client.config.UAAOAuthClient()},
		"client_secret": {client.config.UAAOAuthClientSecret()},
	}

	// An empty grant_type implies that the authentication grant_type is 'password'
	if client.config.UAAGrantType() != "" {
		values.Add("grant_type", client.config.UAAGrantType())
	} else {
		values.Add("grant_type", string(constant.GrantTypeRefreshToken))
		values.Add("refresh_token", refreshToken)
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
