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
	cmd.SharedActor = sharedaction.NewActor()

	ccClient, uaaClient, err := shared.NewClients(config, ui)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, uaaClient)

	return nil
}

func (cmd GetHealthCheckCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(cmd.Config, true, true)
	if err != nil {
		return shared.HandleError(err)
	}

	cmd.UI.DisplayText("Getting health_check_type value for {{.AppName}}",
		map[string]interface{}{
			"AppName": cmd.RequiredArgs.AppName,
		})

	app, warnings, err := cmd.Actor.GetApplicationByNameAndSpace(
		cmd.RequiredArgs.AppName,
		cmd.Config.TargetedSpace().GUID,
	)

	cmd.UI.DisplayWarnings(warnings)

	if err != nil {
		return shared.HandleError(err)
	}

	cmd.UI.DisplayNewline()

	cmd.UI.DisplayText("health_check_type is {{.HealthCheckType}}",
		map[string]interface{}{
			"HealthCheckType": app.HealthCheckType,
		})

	return nil
}
