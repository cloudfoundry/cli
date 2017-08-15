package v2action

import "code.cloudfoundry.org/cli/api/uaa"

//go:generate counterfeiter . UAAClient

type UAAClient interface {
	Authenticate(username string, password string) (string, string, error)
	CreateUser(username string, password string, origin string) (uaa.User, error)
	GetSSHPasscode(accessToken string, sshOAuthClient string) (string, error)
	RefreshAccessToken(refreshToken string) (uaa.RefreshedTokens, error)
}
