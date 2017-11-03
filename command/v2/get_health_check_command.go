package v2

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v2/shared"
)

//go:generate counterfeiter . GetHealthCheckActor
type GetHealthCheckActor interface {
	GetApplicationByNameAndSpace(name string, spaceGUID string) (v2action.Application, v2action.Warnings, error)
}

type GetHealthCheckCommand struct {
	RequiredArgs flag.AppName `positional-args:"yes"`
	usage        interface{}  `usage:"CF_NAME get-health-check APP_NAME"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       GetHealthCheckActor
}

func (cmd *GetHealthCheckCommand) Setup(config command.Config, ui command.UI) error {
	cmd.Config = config
	cmd.UI = ui
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)

	return nil
}

func (cmd GetHealthCheckCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting health check type for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"AppName":   cmd.RequiredArgs.AppName,
			"OrgName":   cmd.Config.TargetedOrganization().Name,
			"SpaceName": cmd.Config.TargetedSpace().Name,
			"Username":  user.Name,
		})

	app, warnings, err := cmd.Actor.GetApplicationByNameAndSpace(
		cmd.RequiredArgs.AppName,
		cmd.Config.TargetedSpace().GUID,
	)

	cmd.UI.DisplayWarnings(warnings)

	if err != nil {
		return err
	}

	cmd.UI.DisplayNewline()

	table := [][]string{
		{cmd.UI.TranslateText("health check type:"), string(app.HealthCheckType)},
		{cmd.UI.TranslateText("endpoint (for http type):"), app.CalculatedHealthCheckEndpoint()},
	}

	cmd.UI.DisplayKeyValueTable("", table, 3)

	return nil
}
