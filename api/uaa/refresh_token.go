package uaa

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"code.cloudfoundry.org/cli/api/uaa/internal"
)

// RefreshToken represents the UAA refresh token response.
type RefreshToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Type         string `json:"token_type"`
}

// AuthorizationToken returns formatted authorization header.
func (refreshTokenResponse RefreshToken) AuthorizationToken() string {
	return fmt.Sprintf("%s %s", refreshTokenResponse.Type, refreshTokenResponse.AccessToken)
}

// RefreshAccessToken refreshes the current access token.
func (client *Client) RefreshAccessToken(refreshToken string) (RefreshToken, error) {
	body := strings.NewReader(url.Values{
		"client_id":     {client.id},
		"client_secret": {client.secret},
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
	}.Encode())

	request, err := client.newRequest(requestOptions{
		RequestName: internal.PostOAuthTokenRequest,
		Header: http.Header{
			"Content-Type": {"application/x-www-form-urlencoded"},
		},
		Body: body,
	})
	if err != nil {
		return RefreshToken{}, err
	}

	var refreshResponse RefreshToken
	response := Response{
		Result: &refreshResponse,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return RefreshToken{}, err
	}

	return refreshResponse, nil
}
