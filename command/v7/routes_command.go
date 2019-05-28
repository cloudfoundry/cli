package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/util/ui"
)

//go:generate counterfeiter . RoutesActor

type RoutesActor interface {
	GetRoutesBySpace(string) ([]v7action.Route, v7action.Warnings, error)
}

type RoutesCommand struct {
	usage           interface{} `usage:"CF_NAME routes"`
	relatedCommands interface{} `related_commands:"check-route, domains, map-route, unmap-route"`

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
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	currentUser, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	targetedOrg := cmd.Config.TargetedOrganization()
	targetedSpace := cmd.Config.TargetedSpace()
	cmd.UI.DisplayTextWithFlavor("Getting routes for org {{.CurrentOrg}} / space {{.CurrentSpace}} as {{.CurrentUser}}...\n", map[string]interface{}{
		"CurrentOrg":   targetedOrg.Name,
		"CurrentSpace": targetedSpace.Name,
		"CurrentUser":  currentUser.Name,
	})

	routes, warnings, err := cmd.Actor.GetRoutesBySpace(targetedSpace.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if len(routes) > 0 {
		cmd.displayRoutesTable(routes)
	} else {
		cmd.UI.DisplayText("No routes found.")
	}
	return nil
}

func (cmd RoutesCommand) displayRoutesTable(routes []v7action.Route) {
	var routesTable = [][]string{
		{
			cmd.UI.TranslateText("space"),
			cmd.UI.TranslateText("host"),
			cmd.UI.TranslateText("domain"),
			cmd.UI.TranslateText("path"),
		},
	}

	for _, route := range routes {
		routesTable = append(routesTable, []string{
			route.SpaceName,
			route.Host,
			route.DomainName,
			route.Path,
		})
	}

	cmd.UI.DisplayTableWithHeader("", routesTable, ui.DefaultTableSpacePadding)
}
