package v2action

import "code.cloudfoundry.org/cli/api/uaa"

//go:generate counterfeiter . UAAClient

type UAAClient interface {
	NewUser(username string, password string, origin string) (uaa.User, error)
}
