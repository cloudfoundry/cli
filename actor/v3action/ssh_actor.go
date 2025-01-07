package v3action

import "code.cloudfoundry.org/cli/v7/actor/sharedaction"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . SSHActor

type SSHActor interface {
	ExecuteSecureShell(sshOptions sharedaction.SSHOptions) error
}
