package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/resources"
)

// CreateUser creates a new Cloud Controller User from the provided UAA user
// ID.
func (client *Client) CreateUser(uaaUserID string) (resources.User, Warnings, error) {
	type userRequestBody struct {
		GUID string `json:"guid"`
	}

	user := userRequestBody{GUID: uaaUserID}
	var responseBody resources.User

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

func (client *Client) GetUser(userGUID string) (resources.User, Warnings, error) {
	var responseBody resources.User

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetUserRequest,
		URIParams:    internal.Params{"user_guid": userGUID},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

func (client *Client) GetUsers(query ...Query) ([]resources.User, Warnings, error) {
	var users []resources.User

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetUsersRequest,
		Query:        query,
		ResponseBody: resources.User{},
		AppendToList: func(item interface{}) error {
			users = append(users, item.(resources.User))
			return nil
		},
	})

	return users, warnings, err
}
