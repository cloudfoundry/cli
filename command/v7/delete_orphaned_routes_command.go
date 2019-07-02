package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/v7/shared"
)

//go:generate counterfeiter . DeleteOrphanedRoutesActor

type DeleteOrphanedRoutesActor interface {
	DeleteOrphanedRoutes(spaceGUID string) (v7action.Warnings, error)
}

type DeleteOrphanedRoutesCommand struct {
	usage           interface{} `usage:"CF_NAME delete-orphaned-routes [-f]\n"`
	Force           bool        `short:"f" description:"Force deletion without confirmation"`
	relatedCommands interface{} `related_commands:"delete-routes, routes"`

	UI          command.UI
	Config      command.Config
	Actor       DeleteOrphanedRoutesActor
	SharedActor command.SharedActor
}

func (cmd *DeleteOrphanedRoutesCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.NewClients(config, ui, true, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, sharedActor, uaaClient)
	return nil
}

func (cmd DeleteOrphanedRoutesCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	_, err = cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	if !cmd.Force {
		response, promptErr := cmd.UI.DisplayBoolPrompt(false, "Really delete orphaned routes?")

		if promptErr != nil {
			return promptErr
		}

		if !response {
			cmd.UI.DisplayText("Routes have not been deleted.")
			return nil
		}
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}
	cmd.UI.DisplayTextWithFlavor("Deleting orphaned routes as {{.userName}}...",
		map[string]interface{}{
			"userName": user.Name,
		})

	warnings, err := cmd.Actor.DeleteOrphanedRoutes(cmd.Config.TargetedSpace().GUID)

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	return nil
}
