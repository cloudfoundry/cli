package v7action

import "code.cloudfoundry.org/cli/api/uaa"

//go:generate counterfeiter . UAAClient

type UAAClient interface {
	GetSSHPasscode(accessToken string, sshOAuthClient string) (string, error)
	GetUsers(userName, origin string) ([]uaa.User, error)
	CreateUser(username string, password string, origin string) (uaa.User, error)
}
