package v7

import (
	"errors"

	"code.cloudfoundry.org/cli/v8/actor/actionerror"
	"code.cloudfoundry.org/cli/v8/command/flag"
)

type UpdateDestinationCommand struct {
	BaseCommand

	RequiredArgs flag.AppDomain   `positional-args:"yes"`
	Hostname     string           `long:"hostname" short:"n" description:"Hostname for the HTTP route (required for shared domains)"`
	AppProtocol  string           `long:"app-protocol" description:"New Protocol for the route destination (http1 or http2). Only applied to HTTP routes"`
	Path         flag.V7RoutePath `long:"path" description:"Path for the HTTP route"`

	relatedCommands interface{} `related_commands:"routes, map-route, create-route, unmap-route"`
}

func (cmd UpdateDestinationCommand) Usage() string {
	return `
Edit an existing HTTP route:
   CF_NAME update-destination APP_NAME DOMAIN [--hostname HOSTNAME] [--app-protocol PROTOCOL] [--path PATH]`
}

func (cmd UpdateDestinationCommand) Examples() string {
	return `
CF_NAME update-destination my-app example.com --hostname myhost --app-protocol http2                   # myhost.example.com
CF_NAME update destination my-app example.com --hostname myhost --path foo --app-protocol http2        # myhost.example.com/foo`
}

func (cmd UpdateDestinationCommand) Execute(args []string) error {
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
	route, warnings, err := cmd.Actor.GetRouteByAttributes(domain, cmd.Hostname, path, 0)
	cmd.UI.DisplayWarnings(warnings)
	url := desiredURL(domain.Name, cmd.Hostname, path, 0)
	if err != nil {
		if _, ok := err.(actionerror.RouteNotFoundError); ok {
			cmd.UI.DisplayText("Route to be updated does not exist.")
			return err
		}
		return err
	}

	dest, err := cmd.Actor.GetRouteDestinationByAppGUID(route, app.GUID)
	if err != nil {
		if _, ok := err.(actionerror.RouteDestinationNotFoundError); !ok {
			cmd.UI.DisplayText("Route's destination to be updated does not exist.")
			return err
		}
	}

	if cmd.AppProtocol == "" {
		cmd.AppProtocol = "http1"
	}

	if cmd.AppProtocol == "tcp" {
		return errors.New("Destination protocol must be 'http1' or 'http2'")
	}

	if dest.Protocol == cmd.AppProtocol {
		cmd.UI.DisplayText(" App '{{ .AppName }}' is already using '{{ .AppProtocol }}'. Nothing has been updated", map[string]interface{}{
			"AppName":     cmd.RequiredArgs.App,
			"AppProtocol": cmd.AppProtocol,
		})
		cmd.UI.DisplayOK()
		return nil
	}

	cmd.UI.DisplayTextWithFlavor("Updating destination protocol from {{.OldProtocol}} to {{.NewProtocol}} for route {{.URL}} in org {{.OrgName}} / space {{.SpaceName}} as {{.User}}...",
		map[string]interface{}{
			"OldProtocol": dest.Protocol,
			"NewProtocol": cmd.AppProtocol,
			"URL":         url,
			"User":        user.Name,
			"SpaceName":   cmd.Config.TargetedSpace().Name,
			"OrgName":     cmd.Config.TargetedOrganization().Name,
		})

	warnings, err = cmd.Actor.UpdateDestination(
		route.GUID,
		dest.GUID,
		cmd.AppProtocol,
	)

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	cmd.UI.DisplayOK()

	return nil
}
