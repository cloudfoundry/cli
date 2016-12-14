package v2action

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"

// User represents a CLI user.
type User ccv2.User

// NewUser creates a new user in UAA and registers it with cloud controller.
func (actor Actor) NewUser(username string, password string, origin string) (User, Warnings, error) {
	uaaUser, err := actor.UAAClient.NewUser(username, password, origin)
	if err != nil {
		return User{}, nil, err
	}

	ccUser, ccWarnings, err := actor.CloudControllerClient.NewUser(uaaUser.ID)

	return User(ccUser), Warnings(ccWarnings), err
}
