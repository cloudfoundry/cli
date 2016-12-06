package v2action

import (
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/uaa"
)

// User represents a CLI user.
type User ccv2.User

// NewUser creates a new user in UAA and registers it with cloud controller.
func (actor Actor) NewUser(username string, password string) (User, Warnings, error) {
	uaaUser, err := actor.UAAClient.NewUser(username, password)
	if err != nil {
		if _, ok := err.(uaa.ConflictError); ok {
			return User{}, Warnings{fmt.Sprintf("user %s already exists", username)}, nil
		}
		return User{}, nil, err
	}

	ccUser, ccWarnings, err := actor.CloudControllerClient.NewUser(uaaUser.ID)

	return User(ccUser), Warnings(ccWarnings), err
}
