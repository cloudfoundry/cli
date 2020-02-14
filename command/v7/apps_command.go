package v7

import (
	"strings"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/util/ui"
)

type AppsCommand struct {
	BaseCommand

	usage           interface{} `usage:"CF_NAME apps [--labels SELECTOR]\n\nEXAMPLES:\n   CF_NAME apps\n   CF_NAME apps --labels 'environment in (production,staging),tier in (backend)'\n   CF_NAME apps --labels 'env=dev,!chargeback-code,tier in (backend,worker)'"`
	relatedCommands interface{} `related_commands:"events, logs, map-route, push, scale, start, stop, restart"`

	Labels string `long:"labels" description:"Selector to filter apps by labels"`
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

	summaries, warnings, err := cmd.Actor.GetAppSummariesForSpace(cmd.Config.TargetedSpace().GUID, cmd.Labels)
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
		table = append(table, []string{
			summary.Name,
			cmd.UI.TranslateText(strings.ToLower(string(summary.State))),
			summary.ProcessSummaries.String(),
			getURLs(summary.Routes),
		})
	}

	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)

	return nil
}

func getURLs(routes []v7action.Route) string {
	var routeURLs []string
	for _, route := range routes {
		routeURLs = append(routeURLs, route.URL)
	}

	return strings.Join(routeURLs, ", ")
}
