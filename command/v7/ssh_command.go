package v7

import (
	"code.cloudfoundry.org/cli/v9/actor/sharedaction"
	"code.cloudfoundry.org/cli/v9/command"
	"code.cloudfoundry.org/cli/v9/command/flag"
	"code.cloudfoundry.org/cli/v9/command/translatableerror"
	"code.cloudfoundry.org/cli/v9/util/clissh"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . SharedSSHActor

type SharedSSHActor interface {
	ExecuteSecureShell(sshClient sharedaction.SecureShellClient, sshOptions sharedaction.SSHOptions) error
}

type SSHCommand struct {
	BaseCommand

	RequiredArgs          flag.AppName             `positional-args:"yes"`
	ProcessIndex          uint                     `long:"app-instance-index" short:"i" default:"0" description:"App process instance index"`
	Commands              []string                 `long:"command" short:"c" description:"Command to run"`
	DisablePseudoTTY      bool                     `long:"disable-pseudo-tty" short:"T" description:"Disable pseudo-tty allocation"`
	ForcePseudoTTY        bool                     `long:"force-pseudo-tty" description:"Force pseudo-tty allocation"`
	LocalPortForwardSpecs []flag.SSHPortForwarding `short:"L" description:"Local port forward specification"`
	ProcessType           string                   `long:"process" default:"web" description:"App process name"`
	RequestPseudoTTY      bool                     `long:"request-pseudo-tty" short:"t" description:"Request pseudo-tty allocation"`
	SkipHostValidation    bool                     `long:"skip-host-validation" short:"k" description:"Skip host key validation. Not recommended!"`
	SkipRemoteExecution   bool                     `long:"skip-remote-execution" short:"N" description:"Do not execute a remote command"`

	usage           interface{} `usage:"CF_NAME ssh APP_NAME [--process PROCESS] [-i INDEX] [-c COMMAND]...\n   [-L [BIND_ADDRESS:]LOCAL_PORT:REMOTE_HOST:REMOTE_PORT]... [--skip-remote-execution]\n   [--disable-pseudo-tty | --force-pseudo-tty | --request-pseudo-tty] [--skip-host-validation]"`
	relatedCommands interface{} `related_commands:"allow-space-ssh, enable-ssh, space-ssh-allowed, ssh-code, ssh-enabled"`
	allproxy        interface{} `environmentName:"all_proxy" environmentDescription:"Specify a proxy server to enable proxying for all requests"`

	SSHActor  SharedSSHActor
	SSHClient *clissh.SecureShell
}

func (cmd *SSHCommand) Setup(config command.Config, ui command.UI) error {
	err := cmd.BaseCommand.Setup(config, ui)
	if err != nil {
		return err
	}

	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor
	cmd.SSHActor = sharedActor
	cmd.SSHClient = clissh.NewDefaultSecureShell()

	return nil
}

func (cmd SSHCommand) Execute(args []string) error {

	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	ttyOption, err := cmd.EvaluateTTYOption()
	if err != nil {
		return err
	}

	var forwardSpecs []sharedaction.LocalPortForward
	for _, spec := range cmd.LocalPortForwardSpecs {
		forwardSpecs = append(forwardSpecs, sharedaction.LocalPortForward(spec))
	}

	sshAuth, warnings, err := cmd.Actor.GetSecureShellConfigurationByApplicationNameSpaceProcessTypeAndIndex(
		cmd.RequiredArgs.AppName,
		cmd.Config.TargetedSpace().GUID,
		cmd.ProcessType,
		cmd.ProcessIndex,
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	err = cmd.SSHActor.ExecuteSecureShell(
		cmd.SSHClient,
		sharedaction.SSHOptions{
			Commands:              cmd.Commands,
			Endpoint:              sshAuth.Endpoint,
			HostKeyFingerprint:    sshAuth.HostKeyFingerprint,
			LocalPortForwardSpecs: forwardSpecs,
			Passcode:              sshAuth.Passcode,
			SkipHostValidation:    cmd.SkipHostValidation,
			SkipRemoteExecution:   cmd.SkipRemoteExecution,
			TTYOption:             ttyOption,
			Username:              sshAuth.Username,
		})
	if err != nil {
		return err
	}

	return nil
}

// EvaluateTTYOption determines which TTY options are mutually exclusive and
// returns an error accordingly.
func (cmd SSHCommand) EvaluateTTYOption() (sharedaction.TTYOption, error) {
	var count int

	option := sharedaction.RequestTTYAuto
	if cmd.DisablePseudoTTY {
		option = sharedaction.RequestTTYNo
		count++
	}
	if cmd.ForcePseudoTTY {
		option = sharedaction.RequestTTYForce
		count++
	}
	if cmd.RequestPseudoTTY {
		option = sharedaction.RequestTTYYes
		count++
	}

	if count > 1 {
		return option, translatableerror.ArgumentCombinationError{Args: []string{
			"--disable-pseudo-tty", "-T", "--force-pseudo-tty", "--request-pseudo-tty", "-t",
		}}
	}

	return option, nil
}
