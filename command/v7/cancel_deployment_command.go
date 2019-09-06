package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . CancelDeploymentActor

type CancelDeploymentActor interface {
	GetApplicationByNameAndSpace(appName string, spaceGUID string) (v7action.Application, v7action.Warnings, error)
	GetLatestActiveDeploymentForApp(appGUID string) (v7action.Deployment, v7action.Warnings, error)
	CancelDeployment(deploymentGUID string) (v7action.Warnings, error)
}

type CancelDeploymentCommand struct {
	RequiredArgs    flag.AppName `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME cancel-deployment APP_NAME\n\nEXAMPLES:\n   cf cancel-deployment my-app"`
	relatedCommands interface{}  `related_commands:"app, push"`

	UI          command.UI
	Config      command.Config
	Actor       CancelDeploymentActor
	SharedActor command.SharedActor
}

func (cmd *CancelDeploymentCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.NewClients(config, ui, true, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, sharedActor, uaaClient, clock.NewClock())
	return nil
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
