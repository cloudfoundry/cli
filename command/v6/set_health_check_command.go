package v6

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"

	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v6/shared"
)

//go:generate counterfeiter . SetHealthCheckActor

type SetHealthCheckActor interface {
	SetApplicationHealthCheckTypeByNameAndSpace(name string, spaceGUID string, healthCheckType constant.ApplicationHealthCheckType, httpEndpoint string) (v2action.Application, v2action.Warnings, error)
}

type SetHealthCheckCommand struct {
	RequiredArgs flag.SetHealthCheckArgs `positional-args:"yes"`
	HTTPEndpoint string                  `long:"endpoint" default:"/" description:"Path on the app"`
	usage        interface{}             `usage:"CF_NAME set-health-check APP_NAME (process | port | http [--endpoint PATH])\n\nTIP: 'none' has been deprecated but is accepted for 'process'.\n\nEXAMPLES:\n   cf set-health-check worker-app process\n   cf set-health-check my-web-app http --endpoint /foo"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       SetHealthCheckActor
}

func (cmd *SetHealthCheckCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd *SetHealthCheckCommand) Execute(args []string) error {
	var err error

	err = cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Updating health check type for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"AppName":   cmd.RequiredArgs.AppName,
			"OrgName":   cmd.Config.TargetedOrganization().Name,
			"SpaceName": cmd.Config.TargetedSpace().Name,
			"Username":  user.Name,
		})

	app, warnings, err := cmd.Actor.SetApplicationHealthCheckTypeByNameAndSpace(
		cmd.RequiredArgs.AppName,
		cmd.Config.TargetedSpace().GUID,
		constant.ApplicationHealthCheckType(cmd.RequiredArgs.HealthCheck.Type),
		cmd.HTTPEndpoint,
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	if app.Started() {
		cmd.UI.DisplayNewline()
		cmd.UI.DisplayText("TIP: An app restart is required for the change to take effect.")
	}

	return nil
}
