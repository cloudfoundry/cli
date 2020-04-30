package v7

import (
	"strings"

	"code.cloudfoundry.org/cli/resources"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/util/ui"
)

type RoutesCommand struct {
	BaseCommand

	usage           interface{} `usage:"CF_NAME routes [--org-level]"`
	relatedCommands interface{} `related_commands:"check-route, create-route, domains, map-route, unmap-route"`
	Orglevel        bool        `long:"org-level" description:"List all the routes for all spaces of current organization"`
	Labels          string      `long:"labels" description:"Selector to filter routes by labels"`
}

func (cmd RoutesCommand) Execute(args []string) error {
	var (
		routes   []resources.Route
		warnings v7action.Warnings
		err      error
	)

	err = cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	currentUser, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	targetedOrg := cmd.Config.TargetedOrganization()
	targetedSpace := cmd.Config.TargetedSpace()

	if cmd.Orglevel {
		cmd.UI.DisplayTextWithFlavor("Getting routes for org {{.CurrentOrg}} as {{.CurrentUser}}...\n", map[string]interface{}{
			"CurrentOrg":  targetedOrg.Name,
			"CurrentUser": currentUser.Name,
		})
		routes, warnings, err = cmd.Actor.GetRoutesByOrg(targetedOrg.GUID, cmd.Labels)
	} else {
		cmd.UI.DisplayTextWithFlavor("Getting routes for org {{.CurrentOrg}} / space {{.CurrentSpace}} as {{.CurrentUser}}...\n", map[string]interface{}{
			"CurrentOrg":   targetedOrg.Name,
			"CurrentSpace": targetedSpace.Name,
			"CurrentUser":  currentUser.Name,
		})
		routes, warnings, err = cmd.Actor.GetRoutesBySpace(targetedSpace.GUID, cmd.Labels)
	}

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	routeSummaries, warnings, err := cmd.Actor.GetRouteSummaries(routes)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if len(routes) > 0 {
		cmd.displayRoutesTable(routeSummaries)
	} else {
		cmd.UI.DisplayText("No routes found.")
	}

	return nil
}

func (cmd RoutesCommand) displayRoutesTable(routeSummaries []v7action.RouteSummary) {
	var routesTable = [][]string{
		{
			cmd.UI.TranslateText("space"),
			cmd.UI.TranslateText("host"),
			cmd.UI.TranslateText("domain"),
			cmd.UI.TranslateText("path"),
			cmd.UI.TranslateText("apps"),
		},
	}

	for _, routeSummary := range routeSummaries {
		routesTable = append(routesTable, []string{
			routeSummary.SpaceName,
			routeSummary.Host,
			routeSummary.DomainName,
			routeSummary.Path,
			strings.Join(routeSummary.AppNames, ", "),
		})
	}

	cmd.UI.DisplayTableWithHeader("", routesTable, ui.DefaultTableSpacePadding)
}
