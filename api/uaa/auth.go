package uaa

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"code.cloudfoundry.org/cli/api/uaa/constant"
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
func (client Client) Authenticate(ID string, secret string, origin string, grantType constant.GrantType) (string, string, error) {
	requestBody := url.Values{
		"grant_type": {string(grantType)},
	}
	switch grantType {
	case constant.GrantTypeClientCredentials:
		requestBody.Set("client_id", ID)
		requestBody.Set("client_secret", secret)
		if origin != "" {
			return "", "", errors.New("Incorrect Usage: The following arguments cannot be used together: --client-credentials, --origin")
		}
	default:
		requestBody.Set("username", ID)
		requestBody.Set("password", secret)
		if origin != "" {
			requestBody.Set("login_hint", "{\"origin\":\""+origin+"\"}")
		}
	}

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

	if grantType == constant.GrantTypePassword {
		request.SetBasicAuth(client.config.UAAOAuthClient(), client.config.UAAOAuthClientSecret())
	}

	responseBody := AuthResponse{}
	response := Response{
		Result: &responseBody,
	}

	err = client.connection.Make(request, &response)
	return responseBody.AccessToken, responseBody.RefreshToken, err
}
