package uaa

import (
	"net/http"
	"net/url"
	"strings"

	"code.cloudfoundry.org/cli/api/uaa/internal"
)

// AuthResponse contains the access token and refresh token which are granted
// after UAA has authorized a user.
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Authenticate sends a username and password to UAA then returns an access
// token and a refresh token.
func (client Client) Authenticate(username string, password string) (string, string, error) {
	requestBody := url.Values{}
	requestBody.Set("username", username)
	requestBody.Set("password", password)
	requestBody.Set("grant_type", "password")

	request, err := client.newRequest(requestOptions{
		RequestName: internal.PostOAuthTokenRequest,
		Header: http.Header{
			"Content-Type": {"application/x-www-form-urlencoded"},
		},
		Body: strings.NewReader(requestBody.Encode()),
	})
	if err != nil {
		return "", "", err
	}
	request.SetBasicAuth(client.id, client.secret)

	responseBody := AuthResponse{}
	response := Response{
		Result: &responseBody,
	}

	err = client.connection.Make(request, &response)
	return responseBody.AccessToken, responseBody.RefreshToken, err
}
