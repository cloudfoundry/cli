package v7

import (
	"code.cloudfoundry.org/cli/v8/actor/actionerror"
	"code.cloudfoundry.org/cli/v8/command/flag"
)

type UnmapRouteCommand struct {
	BaseCommand

	RequiredArgs    flag.AppDomain   `positional-args:"yes"`
	Hostname        string           `long:"hostname" short:"n" description:"Hostname used to identify the HTTP route"`
	Path            flag.V7RoutePath `long:"path" description:"Path used to identify the HTTP route"`
	Port            int              `long:"port" description:"Port used to identify the TCP route"`
	relatedCommands interface{}      `related_commands:"delete-route, map-route, routes"`
}

func (cmd UnmapRouteCommand) Usage() string {
	return `
Unmap an HTTP route:
   CF_NAME unmap-route APP_NAME DOMAIN [--hostname HOSTNAME] [--path PATH]

Unmap a TCP route:
   CF_NAME unmap-route APP_NAME DOMAIN --port PORT`
}

func (cmd UnmapRouteCommand) Examples() string {
	return `
CF_NAME unmap-route my-app example.com                              # example.com
CF_NAME unmap-route my-app example.com --hostname myhost            # myhost.example.com
CF_NAME unmap-route my-app example.com --hostname myhost --path foo # myhost.example.com/foo
CF_NAME unmap-route my-app example.com --port 5000                  # example.com:5000`
}

func (cmd UnmapRouteCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	domain, warnings, err := cmd.Actor.GetDomainByName(cmd.RequiredArgs.Domain)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	spaceGUID := cmd.Config.TargetedSpace().GUID
	app, warnings, err := cmd.Actor.GetApplicationByNameAndSpace(cmd.RequiredArgs.App, spaceGUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	path := cmd.Path.Path
	route, warnings, err := cmd.Actor.GetRouteByAttributes(domain, cmd.Hostname, path, cmd.Port)
	url := desiredURL(domain.Name, cmd.Hostname, path, cmd.Port)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Removing route {{.URL}} from app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.User}}...", map[string]interface{}{
		"URL":       url,
		"AppName":   cmd.RequiredArgs.App,
		"User":      user.Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
	})

	destination, err := cmd.Actor.GetRouteDestinationByAppGUID(route, app.GUID)
	if err != nil {
		if _, ok := err.(actionerror.RouteDestinationNotFoundError); ok {
			cmd.UI.DisplayText("Route to be unmapped is not currently mapped to the application.")
			cmd.UI.DisplayOK()
			return nil
		}

		return err
	}

	warnings, err = cmd.Actor.UnmapRoute(route.GUID, destination.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
