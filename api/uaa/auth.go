package uaa

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
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

type RevokeResponse struct {
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

func (client Client) Revoke(accessToken string) error {

	request1, err1 := client.newRequest(requestOptions{
		RequestName: internal.GetClientList,
		// URIParams:   map[string]string{"user_id": "aed4672c-f06e-4a7c-98a0-c29b8d725c12"},
	})
	request1.Header.Set("Authorization", "Bearer "+accessToken)

	if err1 != nil {
		return err1
	}

	responseBody1 := RevokeResponse{}
	response1 := Response{
		Result: &responseBody1,
	}

	err1 = client.connection.Make(request1, &response1)
	if err1 != nil {
		fmt.Println("what u doin")
	}

	fmt.Printf("\n--%v--\n", string(response1.RawResponse))

	// fmt.Printf("***********8deebug****\n")

	// request, err = client.newRequest(requestOptions{
	// 	RequestName: internal.GetClientList,
	// 	URIParams:   map[string]string{"user_id": "aed4672c-f06e-4a7c-98a0-c29b8d725c12"},
	// 	// URIParams:   "aed4672c-f06e-4a7c-98a0-c29b8d725c12",
	// })
	// request.Header.Set("Authorization", "Bearer "+accessToken)

	// if err != nil {
	// 	return err
	// }

	// responseBody = RevokeResponse{}
	// response = Response{
	// 	Result: &responseBody,
	// }

	// err = client.connection.Make(request, &response)
	// fmt.Println(string(response.RawResponse))

	// return err
	jti, err := client.getJtiFromAccessToken(accessToken)
	if err != nil {
		return err
	}
	request, err := client.newRequest(requestOptions{
		// RequestName: internal.DeleteUserTokenRequest,
		RequestName: internal.DeleteTokenRequest,
		URIParams: map[string]string{
			"token_id": jti,
			// "client_id": "admin",
		},
	})
	request.Header.Set("Authorization", "Bearer "+accessToken)

	if err != nil {
		return err
	}

	responseBody := RevokeResponse{}
	response := Response{
		Result: &responseBody,
	}

	err = client.connection.Make(request, &response)
	return err
}

func (client Client) getJtiFromAccessToken(accessToken string) (string, error) {
	segments := strings.Split(accessToken, ".")

	if len(segments) < 2 {
		return "", errors.New("access token missing segments")
	}

	jsonPayload, err := base64.RawURLEncoding.DecodeString(segments[1])

	if err != nil {
		return "", errors.New("could not base64 decode token payload")
	}

	payload := make(map[string]interface{})
	json.Unmarshal(jsonPayload, &payload)
	jti, ok := payload["jti"].(string)

	if !ok {
		return "", errors.New("could not parse jti from payload")
	}

	return jti, nil
}
