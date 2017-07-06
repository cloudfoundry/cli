package uaa

import (
	"encoding/base64"
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

// oauthClientAuthorizationHeader create a value suitable for use in an authorization header:
// https://tools.ietf.org/html/rfc6749#section-2.3.1
func oauthClientAuthorizationHeader(clientID, clientSecret string) string {
	return fmt.Sprintf("Basic %s",
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s",
			url.QueryEscape(clientID),
			url.QueryEscape(clientSecret)))))
}

// addClientAuthorizationToRequest adds an authorization header for the client ID/secret to the request
func (client *Client) addClientAuthorizationToRequest(req *http.Request) {
	req.Header.Set("Authorization", oauthClientAuthorizationHeader(client.id, client.secret))
}

// RefreshAccessToken refreshes the current access token.
func (client *Client) RefreshAccessToken(refreshToken string) (RefreshToken, error) {
	body := strings.NewReader(url.Values{
		"client_id":     {client.id},     // note, this is doubled up with those added by AddClientAuthorization...
		"client_secret": {client.secret}, // this one too. Should be harmless. Keeping just in case.
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
	}.Encode())

	request, err := client.newRequest(requestOptions{
		RequestName: internal.PostOAuthTokenRequest,
		Header: http.Header{
			"Content-Type": {"application/x-www-form-urlencoded"},
		},
		Body: body,
		AddClientAuthorization: true,
	})
	if err != nil {
		return RefreshToken{}, err
	}
	request.SetBasicAuth(client.id, client.secret)

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
