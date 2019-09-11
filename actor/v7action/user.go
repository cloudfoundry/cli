package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

// User represents a CLI user.
// This means that v7action.User has the same fields as a ccv3.User
type User ccv3.User

// CreateUser creates a new user in UAA and registers it with cloud controller.
func (actor Actor) CreateUser(username string, password string, origin string) (User, Warnings, error) {
	uaaUser, err := actor.UAAClient.CreateUser(username, password, origin)
	if err != nil {
		return User{}, nil, err
	}

	ccUser, ccWarnings, err := actor.CloudControllerClient.CreateUser(uaaUser.ID)

	return User(ccUser), Warnings(ccWarnings), err
}

// GetUser gets a user in UAA with the given username and (if provided) origin.
// It returns an error if no matching user is found.
// It does not handle the potential case where multiple matching users are found.
//   This assumes that the username-origin combination should uniquely identify a user, so multiple
//   matching users should not be returned.
//   However this does not handle the case where "origin" is provided as empty string.
func (actor Actor) GetUser(username, origin string) (User, error) {
	uaaUsers, err := actor.UAAClient.GetUsers(username, origin)
	if err != nil {
		return User{}, err
	}

	if len(uaaUsers) == 0 {
		return User{}, actionerror.UAAUserNotFoundError{}
	}
	uaaUser := uaaUsers[0]

	v7actionUser := User{
		GUID: uaaUser.ID,
	}

	return v7actionUser, nil
}

// DeleteUser
func (actor Actor) DeleteUser(username string, origin string) (Warnings, error) {
	// TODO: when origin is empty, is it nil or ""?
	uaaUser, err := actor.UAAClient.DeleteUser(username, origin)
	if err != nil {
		return nil, err
	}

	ccWarnings, err := actor.CloudControllerClient.DeleteUser(uaaUser.ID)

	return Warnings(ccWarnings), err
}
