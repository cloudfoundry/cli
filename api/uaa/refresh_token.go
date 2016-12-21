package uaa

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"code.cloudfoundry.org/cli/api/uaa/internal"
)

// RefreshTokenResponse represents the UAA refresh token response
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

// AuthorizationToken returns formatted authorization header
func (refreshTokenResponse RefreshTokenResponse) AuthorizationToken() string {
	return fmt.Sprintf("%s %s", refreshTokenResponse.TokenType, refreshTokenResponse.AccessToken)
}

// RefreshToken refreshes the current access token
func (client *Client) RefreshToken() error {
	body := strings.NewReader(url.Values{
		"client_id":     {client.store.UAAOAuthClient()},
		"client_secret": {client.store.UAAOAuthClientSecret()},
		"grant_type":    {"refresh_token"},
		"refresh_token": {client.store.RefreshToken()},
	}.Encode())

	request, err := client.newRequest(requestOptions{
		RequestName: internal.RefreshTokenRequest,
		Header: http.Header{
			"Content-Type": {"application/x-www-form-urlencoded"},
		},
		Body: body,
	})
	if err != nil {
		return err
	}

	var refreshResponse RefreshTokenResponse
	response := Response{
		Result: &refreshResponse,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return err
	}

	client.store.SetAccessToken(refreshResponse.AuthorizationToken())
	client.store.SetRefreshToken(refreshResponse.RefreshToken)
	return nil
}
