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

//go:generate counterfeiter . DisableSSHActor

type DisableSSHActor interface {
	GetApplicationByNameAndSpace(appName string, spaceGUID string) (v7action.Application, v7action.Warnings, error)
	GetAppFeature(appGUID string, featureName string) (ccv3.ApplicationFeature, v7action.Warnings, error)
	UpdateAppFeature(app v7action.Application, enabled bool, featureName string) (v7action.Warnings, error)
}

type DisableSSHCommand struct {
	RequiredArgs    flag.AppName `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME disable-ssh APP_NAME"`
	relatedCommands interface{}  `related_commands:"disallow-space-ssh, space-ssh-allowed, ssh, ssh-enabled"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       DisableSSHActor
}

func (cmd *DisableSSHCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd *DisableSSHCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	username, err := cmd.Config.CurrentUserName()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Disabling ssh support for app {{.AppName}} as {{.CurrentUserName}}...", map[string]interface{}{
		"AppName":         cmd.RequiredArgs.AppName,
		"CurrentUserName": username,
	})

	app, getAppWarnings, err := cmd.Actor.GetApplicationByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)
	if err != nil {
		return err
	}
	cmd.UI.DisplayWarningsV7(getAppWarnings)

	appFeature, getAppFeatureWarnings, err := cmd.Actor.GetAppFeature(app.GUID, "ssh")
	if err != nil {
		return err
	}

	cmd.UI.DisplayWarningsV7(getAppFeatureWarnings)

	if !appFeature.Enabled {
		cmd.UI.DisplayTextWithFlavor("ssh support for app '{{.AppName}}' is already disabled.", map[string]interface{}{
			"AppName": cmd.RequiredArgs.AppName,
		})
	}

	updateSSHWarnings, err := cmd.Actor.UpdateAppFeature(app, false, "ssh")
	if err != nil {
		return err
	}

	cmd.UI.DisplayWarningsV7(updateSSHWarnings)
	cmd.UI.DisplayOK()
	return nil
}
