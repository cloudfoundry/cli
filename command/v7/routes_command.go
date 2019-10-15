package v7

import (
	"strings"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/util/ui"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . RoutesActor

type RoutesActor interface {
	GetRoutesBySpace(string) ([]v7action.Route, v7action.Warnings, error)
	GetRoutesByOrg(string) ([]v7action.Route, v7action.Warnings, error)
	GetRouteSummaries([]v7action.Route) ([]v7action.RouteSummary, v7action.Warnings, error)
}

type RoutesCommand struct {
	usage           interface{} `usage:"CF_NAME routes [--orglevel]"`
	relatedCommands interface{} `related_commands:"check-route, domains, map-route, unmap-route"`
	Orglevel        bool        `long:"orglevel" description:"List all the routes for all spaces of current organization"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       RoutesActor
}

func (cmd *RoutesCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, _, err := shared.GetNewClientsAndConnectToCF(config, ui, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, nil, nil, clock.NewClock())

	return nil
}

func (cmd RoutesCommand) Execute(args []string) error {
	var (
		routes   []v7action.Route
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
		routes, warnings, err = cmd.Actor.GetRoutesByOrg(targetedOrg.GUID)
	} else {
		cmd.UI.DisplayTextWithFlavor("Getting routes for org {{.CurrentOrg}} / space {{.CurrentSpace}} as {{.CurrentUser}}...\n", map[string]interface{}{
			"CurrentOrg":   targetedOrg.Name,
			"CurrentSpace": targetedSpace.Name,
			"CurrentUser":  currentUser.Name,
		})
		routes, warnings, err = cmd.Actor.GetRoutesBySpace(targetedSpace.GUID)
	}

	cmd.UI.DisplayWarningsV7(warnings)
	if err != nil {
		return err
	}

	routeSummaries, warnings, err := cmd.Actor.GetRouteSummaries(routes)
	cmd.UI.DisplayWarningsV7(warnings)
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
