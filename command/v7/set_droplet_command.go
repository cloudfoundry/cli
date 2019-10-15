package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . SetDropletActor

type SetDropletActor interface {
	SetApplicationDropletByApplicationNameAndSpace(appName string, spaceGUID string, dropletGUID string) (v7action.Warnings, error)
}

type SetDropletCommand struct {
	RequiredArgs    flag.AppName `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME set-droplet APP_NAME -d DROPLET_GUID"`
	relatedCommands interface{}  `related_commands:"app, droplets, stage, push, packages, create-package"`

	DropletGUID string `short:"d" long:"droplet-guid" description:"The guid of the droplet to use" required:"true"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       SetDropletActor
}

func (cmd *SetDropletCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, _, err := shared.GetNewClientsAndConnectToCF(config, ui, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, nil, nil, clock.NewClock())

	return nil
}

func (cmd SetDropletCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Setting app {{.AppName}} to droplet {{.DropletGUID}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":     cmd.RequiredArgs.AppName,
		"DropletGUID": cmd.DropletGUID,
		"OrgName":     cmd.Config.TargetedOrganization().Name,
		"SpaceName":   cmd.Config.TargetedSpace().Name,
		"Username":    user.Name,
	})

	warnings, err := cmd.Actor.SetApplicationDropletByApplicationNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID, cmd.DropletGUID)
	cmd.UI.DisplayWarningsV7(warnings)
	if err != nil {
		return err
	}
	cmd.UI.DisplayOK()

	return nil
}
