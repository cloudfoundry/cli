package v2action

import (
	"code.cloudfoundry.org/cli/v7/api/uaa"
	"code.cloudfoundry.org/cli/v7/api/uaa/constant"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . UAAClient

type UAAClient interface {
	APIVersion() string
	Authenticate(credentials map[string]string, origin string, grantType constant.GrantType) (string, string, error)
	CreateUser(username string, password string, origin string) (uaa.User, error)
	GetSSHPasscode(accessToken string, sshOAuthClient string) (string, error)
	LoginPrompts() map[string][]string
	RefreshAccessToken(refreshToken string) (uaa.RefreshedTokens, error)
}
