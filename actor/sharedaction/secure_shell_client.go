package sharedaction

import "code.cloudfoundry.org/cli/util/clissh"

//go:generate counterfeiter . SecureShellClient

type SecureShellClient interface {
	Connect(username string, passcode string, sshEndpoint string, sshHostKeyFingerprint string, skipHostValidation bool) error
	Close() error
	InteractiveSession(commands []string, terminalRequest clissh.TTYRequest) error
	LocalPortForward(localPortForwardSpecs []clissh.LocalPortForward) error
	Wait() error
}
