package v7

import (
	"code.cloudfoundry.org/cli/command/flag"
)

type SetDropletCommand struct {
	command.BaseCommand
	RequiredArgs    flag.AppDroplet `positional-args:"yes"`
	usage           interface{}     `usage:"CF_NAME set-droplet APP_NAME DROPLET_GUID"`
	relatedCommands interface{}     `related_commands:"app, droplets, stage, push, packages, create-package"`
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

	appName := cmd.RequiredArgs.AppName
	dropletGUID := cmd.RequiredArgs.DropletGUID
	org := cmd.Config.TargetedOrganization()
	space := cmd.Config.TargetedSpace()

	cmd.UI.DisplayTextWithFlavor("Setting app {{.AppName}} to droplet {{.DropletGUID}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":     appName,
		"DropletGUID": dropletGUID,
		"OrgName":     org.Name,
		"SpaceName":   space.Name,
		"Username":    user.Name,
	})

	warnings, err := cmd.Actor.SetApplicationDropletByApplicationNameAndSpace(appName, space.GUID, dropletGUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
