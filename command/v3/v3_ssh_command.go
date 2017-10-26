package v3

import (
	"net/http"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v3/shared"
	"code.cloudfoundry.org/cli/util/clissh"
)

//go:generate counterfeiter . V3SSHActor

type V3SSHActor interface {
	CloudControllerAPIVersion() string
	ExecuteSecureShellByApplicationNameSpaceProcessTypeAndIndex(appName string, spaceGUID string, processType string, processIndex uint, sshOptions v3action.SSHOptions) (v3action.Warnings, error)
}

type V3SSHCommand struct {
	RequiredArgs          flag.AppName             `positional-args:"yes"`
	ProcessIndex          uint                     `long:"app-instance-index" short:"i" description:"App process instance index (Default: 0)"`
	Commands              []string                 `long:"command" short:"c" description:"Command to run"`
	DisablePseudoTTY      bool                     `long:"disable-pseudo-tty" short:"T" description:"Disable pseudo-tty allocation"`
	ForcePseudoTTY        bool                     `long:"force-pseudo-tty" description:"Force pseudo-tty allocation"`
	LocalPortForwardSpecs []flag.SSHPortForwarding `short:"L" description:"Local port forward specification"`
	ProcessType           string                   `long:"process" description:"App process name (Default: web)"`
	RequestPseudoTTY      bool                     `long:"request-pseudo-tty" short:"t" description:"Request pseudo-tty allocation"`
	SkipHostValidation    bool                     `long:"skip-host-validation" short:"k" description:"Skip host key validation. Not recommended!"`
	SkipRemoteExecution   bool                     `long:"skip-remote-execution" short:"N" description:"Do not execute a remote command"`

	usage           interface{} `usage:"cf v3-ssh APP_NAME [--process PROCESS] [-i INDEX] [-c COMMAND]...\n   [-L [BIND_ADDRESS:]LOCAL_PORT:REMOTE_HOST:REMOTE_PORT]... [--skip-remote-execution]\n   [--disable-pseudo-tty | --force-pseudo-tty | --request-pseudo-tty] [--skip-host-validation]\n"`
	relatedCommands interface{} `related_commands:"allow-space-ssh, enable-ssh, space-ssh-allowed, ssh-code, ssh-enabled"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       V3SSHActor
}

func (cmd *V3SSHCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config, clissh.NewDefaultSecureShell(ui.GetIn(), ui.GetOut(), ui.GetErr()))
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		if v3Err, ok := err.(ccerror.V3UnexpectedResponseError); ok && v3Err.ResponseCode == http.StatusNotFound {
			return translatableerror.MinimumAPIVersionNotMetError{MinimumVersion: ccversion.MinVersionV3}
		}

		return err
	}

	cmd.Actor = v3action.NewActor(ccClient, config, sharedActor, uaaClient)

	return nil
}

func (cmd V3SSHCommand) Execute(args []string) error {
	err := command.MinimumAPIVersionCheck(cmd.Actor.CloudControllerAPIVersion(), ccversion.MinVersionV3)
	if err != nil {
		return shared.HandleError(err)
	}

	err = cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return shared.HandleError(err)
	}

	ttyOption, err := cmd.evaluateTTYOption()
	if err != nil {
		return shared.HandleError(err)
	}

	if cmd.ProcessType == "" {
		cmd.ProcessType = "web"
	}

	var forwardSpecs []sharedaction.LocalPortForward
	for _, spec := range cmd.LocalPortForwardSpecs {
		forwardSpecs = append(forwardSpecs, sharedaction.LocalPortForward(spec))
	}

	warnings, err := cmd.Actor.ExecuteSecureShellByApplicationNameSpaceProcessTypeAndIndex(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID, cmd.ProcessType, cmd.ProcessIndex, v3action.SSHOptions{
		Commands:              cmd.Commands,
		LocalPortForwardSpecs: forwardSpecs,
		SkipHostValidation:    cmd.SkipHostValidation,
		SkipRemoteExecution:   cmd.SkipRemoteExecution,
		TTYOption:             ttyOption,
	})
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return shared.HandleError(err)
	}

	return nil
}

func (cmd V3SSHCommand) parseForwardSpecs() ([]sharedaction.LocalPortForward, error) {
	return nil, nil
}

// tty options are mutually exclusive
func (cmd V3SSHCommand) evaluateTTYOption() (sharedaction.TTYOption, error) {
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
