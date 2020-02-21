package ccv3

import (
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

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PostUserRequest,
		RequestBody:  user,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

func (client *Client) DeleteUser(uaaUserID string) (JobURL, Warnings, error) {
	jobURL, warnings, err := client.MakeRequest(RequestParams{
		RequestName: internal.DeleteUserRequest,
		URIParams:   internal.Params{"user_guid": uaaUserID},
	})

	return jobURL, warnings, err
}

func (client *Client) GetUser(userGUID string) (User, Warnings, error) {
	var responseBody User

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetUserRequest,
		URIParams:    internal.Params{"user_guid": userGUID},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

func (client *Client) GetUsers(query ...Query) ([]User, Warnings, error) {
	var resources []User

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetUsersRequest,
		Query:        query,
		ResponseBody: User{},
		AppendToList: func(item interface{}) error {
			resources = append(resources, item.(User))
			return nil
		},
	})

	return resources, warnings, err
}
