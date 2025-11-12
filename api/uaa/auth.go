package uaa

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"code.cloudfoundry.org/cli/v8/api/uaa/constant"
	"code.cloudfoundry.org/cli/v8/api/uaa/internal"
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
	} else if grantType == constant.GrantTypeJwtBearer {
		// overwrite client authentication in case of provided parameters in cf auth clientid clientsecret or use defaults as done in password grant
		clientId := client.config.UAAOAuthClient()
		clientSecret := client.config.UAAOAuthClientSecret()
		if creds["client_id"] != "" {
			clientId = creds["client_id"]
		}
		if creds["client_secret"] != "" {
			clientSecret = creds["client_secret"]
		}
		request.SetBasicAuth(clientId, clientSecret)
	}

	responseBody := AuthResponse{}
	response := Response{
		Result: &responseBody,
	}

	err = client.connection.Make(request, &response)
	return responseBody.AccessToken, responseBody.RefreshToken, err
}

func (client Client) Revoke(token string) error {
	jti, err := client.getJtiFromToken(token)
	if err != nil {
		return err
	}

	revokeRequest, err := client.newRequest(requestOptions{
		RequestName: internal.DeleteTokenRequest,
		URIParams: map[string]string{
			"token_id": jti,
		},
	})
	revokeRequest.Header.Set("Authorization", "Bearer "+token)

	if err != nil {
		return err
	}

	err = client.connection.Make(revokeRequest, &Response{})
	return err
}

func (client Client) getJtiFromToken(token string) (string, error) {
	segments := strings.Split(token, ".")

	if len(segments) < 2 {
		return "", errors.New("access token missing segments")
	}

	jsonPayload, err := base64.RawURLEncoding.DecodeString(segments[1])

	if err != nil {
		return "", errors.New("could not base64 decode token payload")
	}

	payload := make(map[string]interface{})
	err = json.Unmarshal(jsonPayload, &payload)
	if err != nil {
		return "", err
	}

	jti, ok := payload["jti"].(string)

	if !ok {
		return "", errors.New("could not parse jti from payload")
	}

	return jti, nil
}
