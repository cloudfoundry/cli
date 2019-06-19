package v7

import (
	"sort"
	"strings"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/util/sorting"
	"code.cloudfoundry.org/cli/util/ui"
)

//go:generate counterfeiter . RoutesActor

type RoutesActor interface {
	GetRoutesBySpace(string) ([]v7action.Route, v7action.Warnings, error)
	GetRoutesByOrg(string) ([]v7action.Route, v7action.Warnings, error)
	GetApplicationsByGUIDs([]string) ([]v7action.Application, v7action.Warnings, error)
	GetRouteDestinations(string) ([]v7action.RouteDestination, v7action.Warnings, error)
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

	ccClient, _, err := shared.NewClients(config, ui, true, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, nil, nil)

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

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	destinations := map[string]string{}

	for _, route := range routes {
		dsts, warnings, err := cmd.Actor.GetRouteDestinations(route.GUID)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}
		appGUIDs := cmd.convertDestinationsToAppGUIDs(dsts)
		apps, warnings, err := cmd.Actor.GetApplicationsByGUIDs(appGUIDs)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}
		destinations[route.GUID] = convertAppsToAppNames(apps)
	}
	sort.Slice(routes, func(i, j int) bool { return sorting.LessIgnoreCase(routes[i].SpaceName, routes[j].SpaceName) })

	if len(routes) > 0 {
		cmd.displayRoutesTable(routes, destinations)
	} else {
		cmd.UI.DisplayText("No routes found.")
	}
	return nil
}

func (cmd RoutesCommand) displayRoutesTable(routes []v7action.Route, destinations map[string]string) {
	var routesTable = [][]string{
		{
			cmd.UI.TranslateText("space"),
			cmd.UI.TranslateText("host"),
			cmd.UI.TranslateText("domain"),
			cmd.UI.TranslateText("path"),
			cmd.UI.TranslateText("apps"),
		},
	}

	for _, route := range routes {
		routesTable = append(routesTable, []string{
			route.SpaceName,
			route.Host,
			route.DomainName,
			route.Path,
			destinations[route.GUID],
		})
	}

	cmd.UI.DisplayTableWithHeader("", routesTable, ui.DefaultTableSpacePadding)
}

func (cmd RoutesCommand) convertDestinationsToAppGUIDs(destinations []v7action.RouteDestination) []string {
	var appGUIDs []string
	for _, dst := range destinations {
		appGUIDs = append(appGUIDs, dst.App.GUID)
	}
	return appGUIDs
}

func convertAppsToAppNames(apps []v7action.Application) string {
	var appNames []string
	for _, app := range apps {
		appNames = append(appNames, app.Name)
	}
	return strings.Join(appNames, ",")
}
