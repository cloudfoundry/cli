package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . SSHEnabledActor

type SSHEnabledActor interface {
	GetSSHEnabledByAppName(appName string, spaceGUID string) (ccv3.SSHEnabled, v7action.Warnings, error)
}

type SSHEnabledCommand struct {
	RequiredArgs    flag.AppName `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME ssh-enabled APP_NAME"`
	relatedCommands interface{}  `related_commands:"enable-ssh, space-ssh-allowed, ssh"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       SSHEnabledActor
}

func (cmd *SSHEnabledCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, _, err := shared.GetNewClientsAndConnectToCF(config, ui, "")
	if err != nil {
		return err
	}

	cmd.Actor = v7action.NewActor(ccClient, config, sharedActor, nil, clock.NewClock())
	return nil
}

func (cmd *SSHEnabledCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	ccv3SSHEnabled, warnings, err := cmd.Actor.GetSSHEnabledByAppName(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if ccv3SSHEnabled.Enabled {
		cmd.UI.DisplayTextWithFlavor("ssh support is enabled for app '{{.AppName}}'.", map[string]interface{}{
			"AppName": cmd.RequiredArgs.AppName,
		})
	} else {
		cmd.UI.DisplayTextWithFlavor("ssh support is disabled for app '{{.AppName}}'.", map[string]interface{}{
			"AppName": cmd.RequiredArgs.AppName,
		})
		cmd.UI.DisplayText(ccv3SSHEnabled.Reason)
	}

	return nil
}
