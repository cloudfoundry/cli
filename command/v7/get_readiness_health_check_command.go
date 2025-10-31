package v7

import (
	"fmt"
	"strconv"

	"code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/command/flag"
	"code.cloudfoundry.org/cli/v9/util/ui"
)

type GetReadinessHealthCheckCommand struct {
	BaseCommand

	RequiredArgs flag.AppName `positional-args:"yes"`
	usage        interface{}  `usage:"CF_NAME get-readiness-health-check APP_NAME"`
}

func (cmd GetReadinessHealthCheckCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting readiness health check type for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   cmd.RequiredArgs.AppName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  user.Name,
	})

	processReadinessHealthChecks, warnings, err := cmd.Actor.GetApplicationProcessReadinessHealthChecksByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayNewline()

	if len(processReadinessHealthChecks) == 0 {
		cmd.UI.DisplayText("App has no processes")
		return nil
	}

	return cmd.DisplayProcessTable(processReadinessHealthChecks)
}

func (cmd GetReadinessHealthCheckCommand) DisplayProcessTable(processReadinessHealthChecks []v7action.ProcessReadinessHealthCheck) error {
	table := [][]string{
		{
			cmd.UI.TranslateText("process"),
			cmd.UI.TranslateText("type"),
			cmd.UI.TranslateText("endpoint (for http)"),
			cmd.UI.TranslateText("invocation timeout"),
			cmd.UI.TranslateText("interval"),
		},
	}

	for _, healthCheck := range processReadinessHealthChecks {
		var invocationTimeout, interval string
		if healthCheck.InvocationTimeout != 0 {
			invocationTimeout = strconv.FormatInt(healthCheck.InvocationTimeout, 10)
		}
		if healthCheck.Interval != 0 {
			interval = strconv.FormatInt(healthCheck.Interval, 10)
		}

		table = append(table, []string{
			healthCheck.ProcessType,
			string(healthCheck.HealthCheckType),
			healthCheck.Endpoint,
			fmt.Sprint(invocationTimeout),
			fmt.Sprint(interval),
		})
	}

	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)

	return nil
}
