package v7action

import (
	"code.cloudfoundry.org/cli/v8/api/uaa"
	"code.cloudfoundry.org/cli/v8/api/uaa/constant"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . UAAClient

type UAAClient interface {
	Authenticate(credentials map[string]string, origin string, grantType constant.GrantType) (string, string, error)
	CreateUser(username string, password string, origin string) (uaa.User, error)
	DeleteUser(userGuid string) (uaa.User, error)
	GetAPIVersion() (string, error)
	GetLoginPrompts() (map[string][]string, error)
	GetSSHPasscode(accessToken string, sshOAuthClient string) (string, error)
	ListUsers(userName, origin string) ([]uaa.User, error)
	RefreshAccessToken(refreshToken string) (uaa.RefreshedTokens, error)
	UpdatePassword(userGUID string, oldPassword string, newPassword string) error
	ValidateClientUser(clientID string) error
	Revoke(token string) error
}
