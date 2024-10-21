package v3action

import "code.cloudfoundry.org/cli/v7/api/uaa/constant"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . UAAClient

type UAAClient interface {
	Authenticate(credentials map[string]string, origin string, grantType constant.GrantType) (string, string, error)
	GetSSHPasscode(accessToken string, sshOAuthClient string) (string, error)
	LoginPrompts() map[string][]string
}
