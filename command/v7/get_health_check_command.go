package v7

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/util/ui"
)

type GetHealthCheckCommand struct {
	BaseCommand

	RequiredArgs flag.AppName `positional-args:"yes"`
	usage        interface{}  `usage:"CF_NAME get-health-check APP_NAME"`
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

	cmd.UI.DisplayTextWithFlavor("Getting health check type for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   cmd.RequiredArgs.AppName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  user.Name,
	})

	processHealthChecks, warnings, err := cmd.Actor.GetApplicationProcessHealthChecksByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayNewline()

	if len(processHealthChecks) == 0 {
		cmd.UI.DisplayText("App has no processes")
		return nil
	}

	return cmd.DisplayProcessTable(processHealthChecks)
}

func (cmd GetHealthCheckCommand) DisplayProcessTable(processHealthChecks []v7action.ProcessHealthCheck) error {
	table := [][]string{
		{
			cmd.UI.TranslateText("process"),
			cmd.UI.TranslateText("health check"),
			cmd.UI.TranslateText("endpoint (for http)"),
			cmd.UI.TranslateText("invocation timeout"),
		},
	}

	for _, healthCheck := range processHealthChecks {
		invocationTimeout := healthCheck.InvocationTimeout
		if invocationTimeout == 0 {
			invocationTimeout = 1
		}

		table = append(table, []string{
			healthCheck.ProcessType,
			string(healthCheck.HealthCheckType),
			healthCheck.Endpoint,
			fmt.Sprint(invocationTimeout),
		})
	}

	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)

	return nil
}
