package ccv3

import (
	ccv3internal "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/api/internal"
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
		RequestName:  ccv3internal.PostUserRequest,
		RequestBody:  user,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

func (client *Client) DeleteUser(uaaUserID string) (JobURL, Warnings, error) {
	jobURL, warnings, err := client.MakeRequest(RequestParams{
		RequestName: ccv3internal.DeleteUserRequest,
		URIParams:   internal.Params{"user_guid": uaaUserID},
	})

	return jobURL, warnings, err
}

func (client *Client) GetUser(userGUID string) (resources.User, Warnings, error) {
	var responseBody resources.User

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  ccv3internal.GetUserRequest,
		URIParams:    internal.Params{"user_guid": userGUID},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

func (client *Client) GetUsers(query ...Query) ([]resources.User, Warnings, error) {
	var users []resources.User

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  ccv3internal.GetUsersRequest,
		Query:        query,
		ResponseBody: resources.User{},
		AppendToList: func(item interface{}) error {
			users = append(users, item.(resources.User))
			return nil
		},
	})

	return users, warnings, err
}

func (client *Client) WhoAmI() (resources.K8sUser, Warnings, error) {
	var user resources.K8sUser

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  ccv3internal.WhoAmI,
		ResponseBody: &user,
	})

	return user, warnings, err
}
