package v7

import (
	"strings"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	v6shared "code.cloudfoundry.org/cli/command/v6/shared"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/util/ui"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . AppsActor

type AppsActor interface {
	GetApplicationsWithProcessesBySpace(spaceGUID string) ([]v7action.ApplicationWithProcessSummary, v7action.Warnings, error)
}

type AppsCommand struct {
	usage           interface{} `usage:"CF_NAME apps"`
	relatedCommands interface{} `related_commands:"events, logs, map-route, push, scale, start, stop, restart"`

	UI          command.UI
	Config      command.Config
	Actor       AppsActor
	V2AppActor  v6shared.V2AppActor
	SharedActor command.SharedActor
}

func (cmd *AppsCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, _, err := shared.NewClients(config, ui, true, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, nil, nil, clock.NewClock())

	ccClientV2, uaaClientV2, err := v6shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}

	cmd.V2AppActor = v2action.NewActor(ccClientV2, uaaClientV2, config)

	return nil
}

func (cmd AppsCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting apps in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  user.Name,
	})
	cmd.UI.DisplayNewline()

	summaries, warnings, err := cmd.Actor.GetApplicationsWithProcessesBySpace(cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if len(summaries) == 0 {
		cmd.UI.DisplayText("No apps found")
		return nil
	}

	table := [][]string{
		{
			cmd.UI.TranslateText("name"),
			cmd.UI.TranslateText("requested state"),
			cmd.UI.TranslateText("processes"),
			cmd.UI.TranslateText("routes"),
		},
	}

	for _, summary := range summaries {
		var routesList string
		if len(summary.ProcessSummaries) > 0 {
			routes, warnings, err := cmd.V2AppActor.GetApplicationRoutes(summary.GUID)
			cmd.UI.DisplayWarnings(warnings)
			if err != nil {
				return err
			}
			routesList = routes.Summary()
		}

		table = append(table, []string{
			summary.Name,
			cmd.UI.TranslateText(strings.ToLower(string(summary.State))),
			summary.ProcessSummaries.String(),
			routesList,
		})
	}

	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)

	return nil
}
