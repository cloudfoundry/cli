package v3

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v3/shared"
	"code.cloudfoundry.org/cli/util/clissh"
)

//go:generate counterfeiter . V3SSHActor

type V3SSHActor interface {
	CloudControllerAPIVersion() string
	ExecuteSecureShellByApplicationNameSpaceProcessTypeAndIndex(appName string, spaceGUID string, processType string, processIndex uint, sshOptions v3action.SSHOptions) (v3action.Warnings, error)
}

type V3SSHCommand struct {
	RequiredArgs flag.AppName `positional-args:"yes"`
	// ProcessType         string       `short:"p" description:"The process name" required:"true"`
	// ProcessIndex        int          `short:"i" description:"The process index" required:"true"`
	// Command             string       `short:"c" description:"command" required:"false"`
	// DisablePseudoTTY    bool         `short:"T" description:"disable pseudo-tty" required:"false"`
	// ForcePseudoTTY      bool         `short:"F" description:"force pseudo-tty" required:"false"`
	// RequestPseudoTTY    bool         `short:"t" description:"request pseudo-tty" required:"false"`
	// Forward             []string     `short:"L" description:"forward" required:"false"`
	// SkipHostValidation  bool         `short:"k" description:"skip host validation" required:"false"`
	// SkipRemoteExecution bool         `short:"N" description:"skip remote execution" required:"false"`

	usage interface{} `usage:"CF_NAME v3-ssh APP_NAME"`

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

	warnings, err := cmd.Actor.ExecuteSecureShellByApplicationNameSpaceProcessTypeAndIndex(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID, "web", 0, v3action.SSHOptions{})
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return shared.HandleError(err)
	}

	return nil
}
