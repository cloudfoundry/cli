package uaa

import (
	"encoding/json"
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
func (client Client) Authenticate(creds map[string]string, origin string, grantType constant.GrantType) (string, string, error) {
	requestBody := url.Values{
		"grant_type": {string(grantType)},
	}

	for k, v := range creds {
		requestBody.Set(k, v)
	}

	type loginHint struct {
		Origin string `json:"origin"`
	}

	originStruct := loginHint{origin}
	originParam, err := json.Marshal(originStruct)
	if err != nil {
		return "", "", err
	}

	var query url.Values
	if origin != "" {
		query = url.Values{
			"login_hint": {string(originParam)},
		}
	}

	request, err := client.newRequest(requestOptions{
		RequestName: internal.PostOAuthTokenRequest,
		Header: http.Header{
			"Content-Type": {"application/x-www-form-urlencoded"},
		},
		Body:  strings.NewReader(requestBody.Encode()),
		Query: query,
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
