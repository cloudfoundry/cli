package v7

import (
	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/command/flag"
)

type DeleteRouteCommand struct {
	BaseCommand

	RequiredArgs    flag.Domain      `positional-args:"yes"`
	Force           bool             `short:"f" description:"Force deletion without confirmation"`
	Hostname        string           `long:"hostname" short:"n" description:"Hostname used to identify the HTTP route (required for shared domains)"`
	Path            flag.V7RoutePath `long:"path" description:"Path used to identify the HTTP route"`
	Port            int              `long:"port" description:"Port used to identify the TCP route"`
	relatedCommands interface{}      `related_commands:"delete-orphaned-routes, routes, unmap-route"`
}

func (cmd DeleteRouteCommand) Usage() string {
	return `
Delete an HTTP route:
   CF_NAME delete-route DOMAIN [--hostname HOSTNAME] [--path PATH] [-f]

Delete a TCP route:
   CF_NAME delete-route DOMAIN --port PORT [-f]`
}

func (cmd DeleteRouteCommand) Examples() string {
	return `
CF_NAME delete-route example.com                              # example.com
CF_NAME delete-route example.com --hostname myhost            # myhost.example.com
CF_NAME delete-route example.com --hostname myhost --path foo # myhost.example.com/foo
CF_NAME delete-route example.com --port 5000                  # example.com:5000`
}

func (cmd DeleteRouteCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	_, err = cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	url := desiredURL(cmd.RequiredArgs.Domain, cmd.Hostname, cmd.Path.Path, cmd.Port)

	cmd.UI.DisplayText("This action impacts all apps using this route.")
	cmd.UI.DisplayText("Deleting this route will make apps unreachable via this route.")

	if !cmd.Force {
		response, promptErr := cmd.UI.DisplayBoolPrompt(false, "Really delete the route {{.URL}}?", map[string]interface{}{
			"URL": url,
		})

		if promptErr != nil {
			return promptErr
		}

		if !response {
			cmd.UI.DisplayText("'{{.URL}}' has not been deleted.", map[string]interface{}{
				"URL": url,
			})
			return nil
		}
	}

	cmd.UI.DisplayTextWithFlavor("Deleting route {{.URL}}...",
		map[string]interface{}{
			"URL": url,
		})

	warnings, err := cmd.Actor.DeleteRoute(cmd.RequiredArgs.Domain, cmd.Hostname, cmd.Path.Path, cmd.Port)

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		if _, ok := err.(actionerror.RouteNotFoundError); ok {
			cmd.UI.DisplayText(`Unable to delete. ` + err.Error())
			cmd.UI.DisplayOK()
			return nil
		}
		return err
	}

	cmd.UI.DisplayOK()
	return nil
}
