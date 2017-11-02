package v3action

import "code.cloudfoundry.org/cli/actor/sharedaction"

//go:generate counterfeiter . SSHActor

type SSHActor interface {
	ExecuteSecureShell(sshOptions sharedaction.SSHOptions) error
}
