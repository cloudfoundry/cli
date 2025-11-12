package sharedaction

import "code.cloudfoundry.org/cli/v8/util/clissh"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . SecureShellClient

type SecureShellClient interface {
	Connect(username string, passcode string, sshEndpoint string, sshHostKeyFingerprint string, skipHostValidation bool) error
	Close() error
	InteractiveSession(commands []string, terminalRequest clissh.TTYRequest) error
	LocalPortForward(localPortForwardSpecs []clissh.LocalPortForward) error
	Wait() error
}
