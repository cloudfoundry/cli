package uaa

import (
	"bytes"
	"encoding/json"
	"net/http"

	"code.cloudfoundry.org/cli/api/uaa/internal"
)

// User represents an UAA user account.
type User struct {
	ID string
}

// newUserRequestBody represents the body of the request.
type newUserRequestBody struct {
	Username string   `json:"userName"`
	Password string   `json:"password"`
	Origin   string   `json:"origin"`
	Name     userName `json:"name"`
	Emails   []email  `json:"emails"`
}

type userName struct {
	FamilyName string `json:"familyName"`
	GivenName  string `json:"givenName"`
}

type email struct {
	Value   string `json:"value"`
	Primary bool   `json:"primary"`
}

// newUserResponse represents the HTTP JSON response.
type newUserResponse struct {
	ID string `json:"id"`
}

// NewUser creates a new UAA user account with the provided password.
func (client *Client) NewUser(user string, password string, origin string) (User, error) {
	userRequest := newUserRequestBody{
		Username: user,
		Password: password,
		Origin:   origin,
		Name: userName{
			FamilyName: user,
			GivenName:  user,
		},
		Emails: []email{
			{
				Value:   user,
				Primary: true,
			},
		},
	}

	bodyBytes, err := json.Marshal(userRequest)
	if err != nil {
		return User{}, err
	}

	request, err := client.newRequest(requestOptions{
		RequestName: internal.NewUserRequest,
		Header: http.Header{
			"Content-Type": {"application/json"},
		},
		Body: bytes.NewBuffer(bodyBytes),
	})
	if err != nil {
		return User{}, err
	}

	var userResponse newUserResponse
	response := Response{
		Result: &userResponse,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return User{}, err
	}

	return User{ID: userResponse.ID}, nil
}
