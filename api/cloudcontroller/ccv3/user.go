package ccv3

import (
	"bytes"
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"encoding/json"
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

func (client *Client) DeleteUser(uaaUserID string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteUserRequest,
		URIParams: internal.Params{
			"user_guid": uaaUserID,
		},
	})

	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}

	err = client.connection.Make(request, &response)

	return response.Warnings, err
}
