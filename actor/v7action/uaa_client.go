package v7action

import "code.cloudfoundry.org/cli/api/uaa"

//go:generate counterfeiter . UAAClient

type UAAClient interface {
	CreateUser(username string, password string, origin string) (uaa.User, error)
	DeleteUser(userGuid string) (uaa.User, error)
	GetSSHPasscode(accessToken string, sshOAuthClient string) (string, error)
	ListUsers(userName, origin string) ([]uaa.User, error)
	RefreshAccessToken(refreshToken string) (uaa.RefreshedTokens, error)
}
