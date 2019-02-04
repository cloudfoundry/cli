package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
)

//go:generate counterfeiter . SetHealthCheckActor

type SetHealthCheckActor interface {
	CloudControllerAPIVersion() string
	SetApplicationProcessHealthCheckTypeByNameAndSpace(appName string, spaceGUID string, healthCheckType constant.HealthCheckType, httpEndpoint string, processType string, invocationTimeout int64) (v7action.Application, v7action.Warnings, error)
}

type SetHealthCheckCommand struct {
	RequiredArgs      flag.SetHealthCheckArgs `positional-args:"yes"`
	HTTPEndpoint      string                  `long:"endpoint" default:"/" description:"Path on the app"`
	InvocationTimeout flag.PositiveInteger    `long:"invocation-timeout" description:"Time (in seconds) that controls individual health check invocations"`
	ProcessType       string                  `long:"process" default:"web" description:"App process to update"`
	usage             interface{}             `usage:"CF_NAME set-health-check APP_NAME (process | port | http [--endpoint PATH]) [--process PROCESS] [--invocation-timeout INVOCATION_TIMEOUT]\n\nEXAMPLES:\n   cf set-health-check worker-app process --process worker\n   cf set-health-check my-web-app http --endpoint /foo\n   cf set-health-check my-web-app http --invocation-timeout 10"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       SetHealthCheckActor
}

func (cmd *SetHealthCheckCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, _, err := shared.NewClients(config, ui, true, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, nil, nil)

	return nil
}

func (cmd SetHealthCheckCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Updating health check type for app {{.AppName}} process {{.ProcessType}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":     cmd.RequiredArgs.AppName,
		"ProcessType": cmd.ProcessType,
		"OrgName":     cmd.Config.TargetedOrganization().Name,
		"SpaceName":   cmd.Config.TargetedSpace().Name,
		"Username":    user.Name,
	})
	cmd.UI.DisplayNewline()

	app, warnings, err := cmd.Actor.SetApplicationProcessHealthCheckTypeByNameAndSpace(
		cmd.RequiredArgs.AppName,
		cmd.Config.TargetedSpace().GUID,
		cmd.RequiredArgs.HealthCheck.Type,
		cmd.HTTPEndpoint,
		cmd.ProcessType,
		cmd.InvocationTimeout.Value,
	)

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	if app.Started() {
		cmd.UI.DisplayText("TIP: An app restart is required for the change to take effect.")
	}

	return nil
}
