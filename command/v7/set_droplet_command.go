package v7

import (
	"code.cloudfoundry.org/cli/command/flag"
)

type SetDropletCommand struct {
	BaseCommand

	RequiredArgs    flag.AppName `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME set-droplet APP_NAME -d DROPLET_GUID"`
	relatedCommands interface{}  `related_commands:"app, droplets, stage, push, packages, create-package"`
	DropletGUID string `short:"d" long:"droplet-guid" description:"The guid of the droplet to use" required:"true"`
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
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	cmd.UI.DisplayOK()

	return nil
}
