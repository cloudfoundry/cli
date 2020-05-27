package v7

import (
	"code.cloudfoundry.org/cli/command/flag"
)

type SetHealthCheckCommand struct {
	BaseCommand

	RequiredArgs      flag.SetHealthCheckArgs `positional-args:"yes"`
	HTTPEndpoint      string                  `long:"endpoint" default:"/" description:"Path on the app"`
	InvocationTimeout flag.PositiveInteger    `long:"invocation-timeout" description:"Time (in seconds) that controls individual health check invocations"`
	ProcessType       string                  `long:"process" default:"web" description:"App process to update"`
	usage             interface{}             `usage:"CF_NAME set-health-check APP_NAME (process | port | http [--endpoint PATH]) [--process PROCESS] [--invocation-timeout INVOCATION_TIMEOUT]\n\nEXAMPLES:\n   cf set-health-check worker-app process --process worker\n   cf set-health-check my-web-app http --endpoint /foo\n   cf set-health-check my-web-app http --invocation-timeout 10"`
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
