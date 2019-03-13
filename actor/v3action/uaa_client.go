package v3action

import "code.cloudfoundry.org/cli/api/uaa/constant"

//go:generate counterfeiter . UAAClient

type UAAClient interface {
	Authenticate(ID string, secret string, origin string, grantType constant.GrantType) (string, string, error)
	GetSSHPasscode(accessToken string, sshOAuthClient string) (string, error)
	LoginPrompts() map[string][]string
}
