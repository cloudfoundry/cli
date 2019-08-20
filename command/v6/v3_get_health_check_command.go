package v6

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v6/shared"
	"code.cloudfoundry.org/cli/util/ui"
)

//go:generate counterfeiter . V3GetHealthCheckActor

type V3GetHealthCheckActor interface {
	GetApplicationProcessHealthChecksByNameAndSpace(appName string, spaceGUID string) ([]v3action.ProcessHealthCheck, v3action.Warnings, error)
}

type V3GetHealthCheckCommand struct {
	RequiredArgs flag.AppName `positional-args:"yes"`
	usage        interface{}  `usage:"CF_NAME v3-get-health-check APP_NAME"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       V3GetHealthCheckActor
}

func (cmd *V3GetHealthCheckCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, _, err := shared.NewV3BasedClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v3action.NewActor(ccClient, config, nil, nil)

	return nil
}

func (cmd V3GetHealthCheckCommand) Execute(args []string) error {
	cmd.UI.DisplayWarning(command.ExperimentalWarning)

	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting process health check types for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   cmd.RequiredArgs.AppName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  user.Name,
	})
	cmd.UI.DisplayNewline()

	processHealthChecks, warnings, err := cmd.Actor.GetApplicationProcessHealthChecksByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if len(processHealthChecks) == 0 {
		cmd.UI.DisplayNewline()
		cmd.UI.DisplayText("App has no processes")
		return nil
	}

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
