package uaa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

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

type paginatedUsersResponse struct {
	Resources []newUserResponse `json:"resources"`
}

type listUsersResponse struct {
	Resources []User `json:resources`
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

// GetUsers gets a list of users from UAA with the given username and (if provided) origin.
// NOTE: that this is a paginated response and we are only currently returning the first page
// of users. This will mean, if no origin is passed and there are more than 100 users with
// the given username, only the first 100 will be returned. For our current purposes, this is
// more than enough, but it would be a problem if we ever need to get all users with a username.
func (client Client) GetUsers(userName, origin string) ([]User, error) {
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

func (client *Client) DeleteUser(user string, origin string) (User, error) {
	if origin == "" {
		origin = "uaa"
	}

	listRequest, err := client.newRequest(requestOptions{
		RequestName: internal.ListUsersRequest,
		Header: http.Header{
			"Content-Type": {"application/json"},
		},
		Query: url.Values{
			"filter": {fmt.Sprintf("username eq \"%s\" and origin eq \"%s\"", user, origin)},
		},
	})

	if err != nil {
		return User{}, err
	}

	var listUsersResponse listUsersResponse
	listResponse := Response{
		Result: &listUsersResponse,
	}

	err = client.connection.Make(listRequest, &listResponse)
	if err != nil {
		return User{}, err
	}
	//TODO: what happens when we get multiple users back from UAA or
	// no user is returned == ENOTFOUND
	deleteRequest, err := client.newRequest(requestOptions{
		RequestName: internal.DeleteUserRequest,
		Header: http.Header{
			"Content-Type": {"application/json"},
		},
		URIParams: map[string]string{"user_guid": User(listUsersResponse.Resources[0]).ID},
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
