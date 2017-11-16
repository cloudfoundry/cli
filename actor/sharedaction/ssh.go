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

func (actor Actor) ExecuteSecureShell(sshClient SecureShellClient, sshOptions SSHOptions) error {
	err := sshClient.Connect(sshOptions.Username, sshOptions.Passcode, sshOptions.Endpoint, sshOptions.HostKeyFingerprint, sshOptions.SkipHostValidation)
	if err != nil {
		return err
	}
	defer sshClient.Close()

	err = sshClient.LocalPortForward(convertActorToSSHPackageForwardingSpecs(sshOptions.LocalPortForwardSpecs))
	if err != nil {
		return err
	}

	if sshOptions.SkipRemoteExecution {
		err = sshClient.Wait()
	} else {
		err = sshClient.InteractiveSession(sshOptions.Commands, clissh.TTYRequest(sshOptions.TTYOption))
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
