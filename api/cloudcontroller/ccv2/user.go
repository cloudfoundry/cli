package ccv2

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/v7/api/cloudcontroller"
	"code.cloudfoundry.org/cli/v7/api/cloudcontroller/ccv2/internal"
)

// User represents a Cloud Controller User.
type User struct {
	// GUID is the unique user identifier.
	GUID string
}

// UnmarshalJSON helps unmarshal a Cloud Controller User response.
func (user *User) UnmarshalJSON(data []byte) error {
	var ccUser struct {
		Metadata internal.Metadata `json:"metadata"`
	}
	err := cloudcontroller.DecodeJSON(data, &ccUser)
	if err != nil {
		return err
	}

	user.GUID = ccUser.Metadata.GUID
	return nil
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
	if err != nil {
		return User{}, response.Warnings, err
	}

	return user, response.Warnings, nil
}
