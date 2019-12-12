package uaa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/uaa/internal"
)

// User represents an UAA user account.
type User struct {
	ID     string
	Origin string
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
	ID     string `json:"id"`
	Origin string `json:"origin"`
}

type paginatedUsersResponse struct {
	Resources []newUserResponse `json:"resources"`
}

// CreateUser creates a new UAA user account with the provided password.
func (client *Client) CreateUser(user string, password string, origin string) (User, error) {
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
		RequestName: internal.PostUserRequest,
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

	return User(userResponse), nil
}

func (client *Client) DeleteUser(userGuid string) (User, error) {
	deleteRequest, err := client.newRequest(requestOptions{
		RequestName: internal.DeleteUserRequest,
		Header: http.Header{
			"Content-Type": {"application/json"},
		},
		URIParams: map[string]string{"user_guid": userGuid},
	})

	if err != nil {
		return User{}, err
	}

	var deleteUserResponse newUserResponse
	deleteResponse := Response{
		Result: &deleteUserResponse,
	}

	err = client.connection.Make(deleteRequest, &deleteResponse)
	if err != nil {
		return User{}, err
	}

	return User(deleteUserResponse), nil
}

// ListUsers gets a list of users from UAA with the given username and (if provided) origin.
// NOTE: that this is a paginated response and we are only currently returning the first page
// of users. This will mean, if no origin is passed and there are more than 100 users with
// the given username, only the first 100 will be returned. For our current purposes, this is
// more than enough, but it would be a problem if we ever need to get all users with a username.
func (client Client) ListUsers(userName, origin string) ([]User, error) {
	filter := fmt.Sprintf(`userName eq "%s"`, userName)

	if origin != "" {
		filter = fmt.Sprintf(`%s and origin eq "%s"`, filter, origin)
	}

	request, err := client.newRequest(requestOptions{
		RequestName: internal.ListUsersRequest,
		Header: http.Header{
			"Content-Type": {"application/json"},
		},
		Query: url.Values{
			"filter": {filter},
		},
	})
	if err != nil {
		return nil, err
	}

	var usersResponse paginatedUsersResponse
	response := Response{
		Result: &usersResponse,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return nil, err
	}

	var users []User
	for _, user := range usersResponse.Resources {
		users = append(users, User(user))
	}

	return users, nil
}

func (client Client) ValidateClientUser(clientID string) error {
	request, err := client.newRequest(requestOptions{
		RequestName: internal.GetClientUser,
		Header: http.Header{
			"Content-Type": {"application/json"},
		},
		URIParams: map[string]string{"client_id": clientID},
	})
	if err != nil {
		return err
	}
	err = client.connection.Make(request, &Response{})

	if errType, ok := err.(RawHTTPStatusError); ok {
		switch errType.StatusCode {
		case http.StatusNotFound:
			return actionerror.UserNotFoundError{Username: clientID}
		case http.StatusForbidden:
			return InsufficientScopeError{}
		}
	}

	return err
}
