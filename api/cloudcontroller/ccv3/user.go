package ccv3

import (
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

	user := userRequestBody{GUID: uaaUserID}
	var responseBody User

	_, warnings, err := client.makeRequest(requestParams{
		RequestName:  internal.PostUserRequest,
		RequestBody:  user,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

func (client *Client) DeleteUser(uaaUserID string) (JobURL, Warnings, error) {
	return client.makeRequest(requestParams{
		RequestName: internal.DeleteUserRequest,
		URIParams:   internal.Params{"user_guid": uaaUserID},
	})
}

func (client *Client) GetUser(userGUID string) (User, Warnings, error) {
	var responseBody User

	_, warnings, err := client.makeRequest(requestParams{
		RequestName:  internal.GetUserRequest,
		URIParams:    internal.Params{"user_guid": userGUID},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
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
