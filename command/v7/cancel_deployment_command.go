package v7

import (
	"code.cloudfoundry.org/cli/command/flag"
)

type CancelDeploymentCommand struct {
	BaseCommand

	RequiredArgs    flag.AppName `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME cancel-deployment APP_NAME\n\nEXAMPLES:\n   cf cancel-deployment my-app"`
	relatedCommands interface{}  `related_commands:"app, push"`
}

func (cmd *CancelDeploymentCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	userName, err := cmd.Config.CurrentUserName()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor(
		"Canceling deployment for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.UserName}}...",
		map[string]interface{}{
			"AppName":   cmd.RequiredArgs.AppName,
			"OrgName":   cmd.Config.TargetedOrganization().Name,
			"SpaceName": cmd.Config.TargetedSpace().Name,
			"UserName":  userName,
		},
	)

	application, warnings, err := cmd.Actor.GetApplicationByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	deployment, warnings, err := cmd.Actor.GetLatestActiveDeploymentForApp(application.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	warnings, err = cmd.Actor.CancelDeployment(deployment.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayText("TIP: Run 'cf app {{.AppName}}' to view app status.", map[string]interface{}{"AppName": cmd.RequiredArgs.AppName})
	return nil
}
