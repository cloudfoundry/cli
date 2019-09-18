package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
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
	uaaUsers, err := actor.UAAClient.ListUsers(username, origin)
	if err != nil {
		return User{}, err
	}

	if len(uaaUsers) == 0 {
		return User{}, actionerror.UAAUserNotFoundError{Username: username}
	}
	if len(uaaUsers) > 1 {
		var origins []string
		for _, user := range uaaUsers {
			origins = append(origins, user.Origin)
		}
		return User{}, actionerror.MultipleUAAUsersFoundError{Username: username, Origins: origins}
	}

	uaaUser := uaaUsers[0]

	v7actionUser := User{
		GUID: uaaUser.ID,
	}
	return v7actionUser, nil
}

// DeleteUser
func (actor Actor) DeleteUser(userGuid string) (Warnings, error) {
	ccWarnings, err := actor.CloudControllerClient.DeleteUser(userGuid)

	if _, ok := err.(ccerror.ResourceNotFoundError); !ok && err != nil {
		return nil, err
	}

	_, err = actor.UAAClient.DeleteUser(userGuid)

	return Warnings(ccWarnings), err
}
