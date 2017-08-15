package uaa

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

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
	body := strings.NewReader(url.Values{
		"client_id":     {client.id},
		"client_secret": {client.secret},
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
	}.Encode())

	request, err := client.newRequest(requestOptions{
		RequestName: internal.PostOAuthTokenRequest,
		Header:      http.Header{"Content-Type": {"application/x-www-form-urlencoded"}},
		Body:        body,
	})
	if err != nil {
		return RefreshedTokens{}, err
	}

	request.SetBasicAuth(client.id, client.secret)

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
