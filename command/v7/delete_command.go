package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/command/flag"
)

type DeleteCommand struct {
	command.BaseCommand

	RequiredArgs       flag.AppName `positional-args:"yes"`
	Force              bool         `short:"f" description:"Force deletion without confirmation"`
	DeleteMappedRoutes bool         `short:"r" description:"Also delete any mapped routes"`
	usage              interface{}  `usage:"CF_NAME delete APP_NAME [-r] [-f]"`
	relatedCommands    interface{}  `related_commands:"apps, scale, stop"`
}

func (cmd DeleteCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	currentUser, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	if !cmd.Force {
		prompt := "Really delete the app {{.AppName}}?"
		if cmd.DeleteMappedRoutes {
			prompt = "Really delete the app {{.AppName}} and associated routes?"
			cmd.UI.DisplayText("Deleting the app and associated routes will make apps with this route, in any org, unreachable.")
		}

		response, promptErr := cmd.UI.DisplayBoolPrompt(false, prompt, map[string]interface{}{
			"AppName": cmd.RequiredArgs.AppName,
		})

		if promptErr != nil {
			return promptErr
		}

		if !response {
			cmd.UI.DisplayText("App '{{.AppName}}' has not been deleted.", map[string]interface{}{
				"AppName": cmd.RequiredArgs.AppName,
			})
			return nil
		}
	}

	cmd.UI.DisplayTextWithFlavor("Deleting app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   cmd.RequiredArgs.AppName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  currentUser.Name,
	})

	warnings, err := cmd.Actor.DeleteApplicationByNameAndSpace(
		cmd.RequiredArgs.AppName,
		cmd.Config.TargetedSpace().GUID,
		cmd.DeleteMappedRoutes,
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		switch err.(type) {
		case actionerror.ApplicationNotFoundError:
			cmd.UI.DisplayWarning("App '{{.AppName}}' does not exist.", map[string]interface{}{
				"AppName": cmd.RequiredArgs.AppName,
			})
		case actionerror.RouteBoundToMultipleAppsError:
			cmd.UI.DeferText(
				"\nTIP: Run 'cf delete {{.AppName}}' to delete the app and 'cf delete-route' to delete the route.",
				map[string]interface{}{
					"AppName": cmd.RequiredArgs.AppName,
				},
			)
			return err
		default:
			return err
		}
	}

	cmd.UI.DisplayOK()

	return nil
}
