package ccv3

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

// User represents a Cloud Controller User.
type User struct {
	// GUID is the unique user identifier.
	GUID             string `json:"guid"`
	Username         string `json:"username"`
	PresentationName string `json:"presentation_name"`
	Origin           string `json:"origin"`
}

// CreateUser creates a new Cloud Controller User from the provided UAA user
// ID.
func (client *Client) CreateUser(uaaUserID string) (User, Warnings, error) {
	type userRequestBody struct {
		GUID string `json:"guid"`
	}

	bodyBytes, err := json.Marshal(userRequestBody{
		GUID: uaaUserID,
	})
	if err != nil {
		return User{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostUserRequest,
		Body:        bytes.NewReader(bodyBytes),
	})
	if err != nil {
		return User{}, nil, err
	}

	var user User
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &user,
	}

	err = client.connection.Make(request, &response)

	return user, response.Warnings, err
}

func (client *Client) DeleteUser(uaaUserID string) (JobURL, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteUserRequest,
		URIParams: internal.Params{
			"user_guid": uaaUserID,
		},
	})

	if err != nil {
		return JobURL(""), nil, err
	}

	response := cloudcontroller.Response{}

	err = client.connection.Make(request, &response)

	return JobURL(response.ResourceLocationURL), response.Warnings, err
}

func (client *Client) GetUser(userGUID string) (User, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetUserRequest,
		URIParams:   map[string]string{"user_guid": userGUID},
	})
	if err != nil {
		return User{}, nil, err
	}

	var responseUser User
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &responseUser,
	}
	err = client.connection.Make(request, &response)

	return responseUser, response.Warnings, err
}

func (client *Client) GetUsers(query ...Query) ([]User, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetUsersRequest,
		Query:       query,
	})
	if err != nil {
		return nil, nil, err
	}

	var usersList []User
	warnings, err := client.paginate(request, User{}, func(item interface{}) error {
		if user, ok := item.(User); ok {
			usersList = append(usersList, user)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   User{},
				Unexpected: item,
			}
		}
		return nil
	})

	return usersList, warnings, err
}
