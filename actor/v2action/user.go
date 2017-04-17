package v2action

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"

// User represents a CLI user.
type User ccv2.User

// CreateUser creates a new user in UAA and registers it with cloud controller.
func (actor Actor) CreateUser(username string, password string, origin string) (User, Warnings, error) {
	uaaUser, err := actor.UAAClient.CreateUser(username, password, origin)
	if err != nil {
		return User{}, nil, err
	}

	ccUser, ccWarnings, err := actor.CloudControllerClient.CreateUser(uaaUser.ID)

	return User(ccUser), Warnings(ccWarnings), err
}
