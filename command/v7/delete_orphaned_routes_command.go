package v7

type DeleteOrphanedRoutesCommand struct {
	BaseCommand

	usage           interface{} `usage:"CF_NAME delete-orphaned-routes [-f]\n"`
	Force           bool        `short:"f" description:"Force deletion without confirmation"`
	relatedCommands interface{} `related_commands:"delete-routes, routes"`
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
