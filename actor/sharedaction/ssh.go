package sharedaction

import "code.cloudfoundry.org/cli/util/clissh"

type TTYOption clissh.TTYRequest

const (
	RequestTTYAuto TTYOption = iota
	RequestTTYNo
	RequestTTYYes
	RequestTTYForce
)

type LocalPortForward clissh.LocalPortForward

type SSHOptions struct {
	Commands              []string
	Username              string
	Passcode              string
	Endpoint              string
	HostKeyFingerprint    string
	SkipHostValidation    bool
	SkipRemoteExecution   bool
	TTYOption             TTYOption
	LocalPortForwardSpecs []LocalPortForward
}

func (actor Actor) ExecuteSecureShell(sshOptions SSHOptions) error {
	err := actor.SecureShellClient.Connect(sshOptions.Username, sshOptions.Passcode, sshOptions.Endpoint, sshOptions.HostKeyFingerprint, false)
	defer actor.SecureShellClient.Close()
	if err != nil {
		return err
	}

	err = actor.SecureShellClient.LocalPortForward(convertActorToSSHPackageForwardingSpecs(sshOptions.LocalPortForwardSpecs))

	if sshOptions.SkipRemoteExecution {
		err = actor.SecureShellClient.Wait()
	} else {
		err = actor.SecureShellClient.InteractiveSession(sshOptions.Commands, clissh.TTYRequest(sshOptions.TTYOption))
	}
	return err
}

func convertActorToSSHPackageForwardingSpecs(actorSpecs []LocalPortForward) []clissh.LocalPortForward {
	sshPackageSpecs := []clissh.LocalPortForward{}

	for _, spec := range actorSpecs {
		sshPackageSpecs = append(sshPackageSpecs, clissh.LocalPortForward(spec))
	}

	return sshPackageSpecs
}
