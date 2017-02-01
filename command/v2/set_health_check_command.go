package v2

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"

	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v2/shared"
)

//go:generate counterfeiter . SetHealthCheckActor
type SetHealthCheckActor interface {
	SetApplicationHealthCheckTypeByNameAndSpace(name string, spaceGUID string, healthCheckType string) (v2action.Warnings, error)
}

type SetHealthCheckCommand struct {
	RequiredArgs flag.SetHealthCheckArgs `positional-args:"yes"`
	usage        interface{}             `usage:"CF_NAME set-health-check APP_NAME ('port' | 'none')"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       SetHealthCheckActor
}

func (cmd *SetHealthCheckCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd *SetHealthCheckCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(cmd.Config, true, true)
	if err != nil {
		return shared.HandleError(err)
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Updating app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"AppName":   cmd.RequiredArgs.AppName,
			"OrgName":   cmd.Config.TargetedOrganization().Name,
			"SpaceName": cmd.Config.TargetedSpace().Name,
			"Username":  user.Name,
		})

	warnings, err := cmd.Actor.SetApplicationHealthCheckTypeByNameAndSpace(
		cmd.RequiredArgs.AppName,
		cmd.Config.TargetedSpace().GUID,
		cmd.RequiredArgs.HealthCheck.Type,
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return shared.HandleError(err)
	}

	cmd.UI.DisplayOK()
	return nil
}
