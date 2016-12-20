package v2

import (
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/v2/shared"
)

//go:generate counterfeiter . DeleteOrphanedRoutesActor

type DeleteOrphanedRoutesActor interface {
	GetOrphanedRoutesBySpace(spaceGUID string) ([]v2action.Route, v2action.Warnings, error)
	DeleteRoute(routeGUID string) (v2action.Warnings, error)
}

type DeleteOrphanedRoutesCommand struct {
	Force           bool        `short:"f" description:"Force deletion without confirmation"`
	usage           interface{} `usage:"CF_NAME delete-orphaned-routes [-f]"`
	relatedCommands interface{} `related_commands:"delete-route, routes"`

	UI     command.UI
	Actor  DeleteOrphanedRoutesActor
	Config command.Config
}

func (cmd *DeleteOrphanedRoutesCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config

	client, _, err := shared.NewClients(config, ui)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(client, nil)

	return nil
}

func (cmd *DeleteOrphanedRoutesCommand) Execute(args []string) error {
	err := command.CheckTarget(cmd.Config, true, true)
	if err != nil {
		return err
	}

	if !cmd.Force {
		deleteOrphanedRoutes, promptErr := cmd.UI.DisplayBoolPrompt("Really delete orphaned routes?", false)
		if promptErr != nil {
			return promptErr
		}

		if !deleteOrphanedRoutes {
			return nil
		}
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting routes as {{.CurrentUser}} ...", map[string]interface{}{
		"CurrentUser": user.Name,
	})
	cmd.UI.DisplayNewline()

	routes, warnings, err := cmd.Actor.GetOrphanedRoutesBySpace(cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		switch err.(type) {
		case v2action.OrphanedRoutesNotFoundError:
		// Do nothing to parity the existing behavior
		default:
			return shared.HandleError(err)
		}
	}

	for _, route := range routes {
		cmd.UI.DisplayText("Deleting route {{.Route}} ...", map[string]interface{}{
			"Route": route.String(),
		})

		warnings, err = cmd.Actor.DeleteRoute(route.GUID)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return shared.HandleError(err)
		}
	}

	cmd.UI.DisplayOK()

	return nil
}
