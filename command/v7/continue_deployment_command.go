package v7

import (
	"code.cloudfoundry.org/cli/v9/command/flag"
	"code.cloudfoundry.org/cli/v9/command/v7/shared"
)

type ContinueDeploymentCommand struct {
	BaseCommand

	RequiredArgs    flag.AppName `positional-args:"yes"`
	NoWait          bool         `long:"no-wait" description:"Exit when the first instance of the web process is healthy"`
	usage           interface{}  `usage:"CF_NAME continue-deployment APP_NAME [--no-wait]\n\nEXAMPLES:\n   cf continue-deployment my-app"`
	relatedCommands interface{}  `related_commands:"app, push"`
}

func (cmd *ContinueDeploymentCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor(
		"Continuing deployment for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.UserName}}...\n",
		map[string]interface{}{
			"AppName":   cmd.RequiredArgs.AppName,
			"OrgName":   cmd.Config.TargetedOrganization().Name,
			"SpaceName": cmd.Config.TargetedSpace().Name,
			"UserName":  user.Name,
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

	warnings, err = cmd.Actor.ContinueDeployment(deployment.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayText("Waiting for app to deploy...\n")

	handleInstanceDetails := func(instanceDetails string) {
		cmd.UI.DisplayText(instanceDetails)
	}

	warnings, err = cmd.Actor.PollStartForDeployment(application, deployment.GUID, cmd.NoWait, handleInstanceDetails)
	cmd.UI.DisplayNewline()
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	summary, warnings, err := cmd.Actor.GetDetailedAppSummary(application.Name, application.SpaceGUID, false)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	appSummaryDisplayer := shared.NewAppSummaryDisplayer(cmd.UI)
	appSummaryDisplayer.AppDisplay(summary, false)

	cmd.UI.DisplayText("\nTIP: Run 'cf app {{.AppName}}' to view app status.", map[string]interface{}{"AppName": cmd.RequiredArgs.AppName})
	return nil
}
