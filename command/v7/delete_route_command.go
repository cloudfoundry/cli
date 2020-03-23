package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/command/flag"
)

type DeleteRouteCommand struct {
	BaseCommand

	RequiredArgs    flag.Domain      `positional-args:"yes"`
	usage           interface{}      `usage:"CF_NAME delete-route DOMAIN [--hostname HOSTNAME] [--path PATH] [-f]\n\nEXAMPLES:\n   CF_NAME delete-route example.com                             # example.com\n   CF_NAME delete-route example.com --hostname myhost            # myhost.example.com\n   CF_NAME delete-route example.com --hostname myhost --path foo # myhost.example.com/foo"`
	Force           bool             `short:"f" description:"Force deletion without confirmation"`
	Hostname        string           `long:"hostname" short:"n" description:"Hostname used to identify the HTTP route (required for shared domains)"`
	Path            flag.V7RoutePath `long:"path" description:"Path used to identify the HTTP route"`
	relatedCommands interface{}      `related_commands:"delete-orphaned-routes, routes, unmap-route"`
}

func (cmd DeleteRouteCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	_, err = cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	domain := cmd.RequiredArgs.Domain
	hostname := cmd.Hostname
	pathName := cmd.Path.Path
	fqdn := desiredFQDN(domain, hostname, pathName)

	cmd.UI.DisplayText("This action impacts all apps using this route.")
	cmd.UI.DisplayText("Deleting this route will make apps unreachable via this route.")

	if !cmd.Force {
		response, promptErr := cmd.UI.DisplayBoolPrompt(false, "Really delete the route {{.FQDN}}?", map[string]interface{}{
			"FQDN": fqdn,
		})

		if promptErr != nil {
			return promptErr
		}

		if !response {
			cmd.UI.DisplayText("'{{.FQDN}}' has not been deleted.", map[string]interface{}{
				"FQDN": fqdn,
			})
			return nil
		}
	}

	cmd.UI.DisplayTextWithFlavor("Deleting route {{.FQDN}}...",
		map[string]interface{}{
			"FQDN": fqdn,
		})

	warnings, err := cmd.Actor.DeleteRoute(domain, hostname, pathName)

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
