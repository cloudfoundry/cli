package ccv2

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// User represents a Cloud Controller User.
type User struct {
	GUID string
}

// userRequestBody represents the body of the request.
type userRequestBody struct {
	GUID string `json:"guid"`
}

// UnmarshalJSON helps unmarshal a Cloud Controller User response.
func (user *User) UnmarshalJSON(data []byte) error {
	var ccUser struct {
		Metadata internal.Metadata `json:"metadata"`
	}
	if err := json.Unmarshal(data, &ccUser); err != nil {
		return err
	}

	user.GUID = ccUser.Metadata.GUID
	return nil
}

// NewUser creates a new Cloud Controller User from the provided UAA user ID.
func (client *Client) NewUser(uaaUserID string) (User, Warnings, error) {
	bodyBytes, err := json.Marshal(userRequestBody{
		GUID: uaaUserID,
	})
	if err != nil {
		return User{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.UsersRequest,
		Body:        bytes.NewBuffer(bodyBytes),
	})
	if err != nil {
		return User{}, nil, err
	}

	var user User
	response := cloudcontroller.Response{
		Result: &user,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return User{}, response.Warnings, err
	}

	return user, response.Warnings, nil
}
