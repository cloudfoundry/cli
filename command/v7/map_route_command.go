package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/command/flag"
)

type MapRouteCommand struct {
	BaseCommand

	RequiredArgs    flag.AppDomain   `positional-args:"yes"`
	usage           interface{}      `usage:"Map an HTTP route:\n      CF_NAME map-route APP_NAME DOMAIN [--hostname HOSTNAME] [--path PATH]\n\n   Map a TCP route:\n      CF_NAME map-route APP_NAME DOMAIN [--port PORT]\n\nEXAMPLES:\n   CF_NAME map-route my-app example.com                              # example.com\n   CF_NAME map-route my-app example.com --hostname myhost            # myhost.example.com\n   CF_NAME map-route my-app example.com --hostname myhost --path foo # myhost.example.com/foo\n   CF_NAME map-route my-app example.com --port 5000                  # example.com:5000"`
	Hostname        string           `long:"hostname" short:"n" description:"Hostname for the HTTP route (required for shared domains)"`
	Path            flag.V7RoutePath `long:"path" description:"Path for the HTTP route"`
	Port            int              `long:"port" description:"Port for the TCP route (default: random port)"`
	relatedCommands interface{}      `related_commands:"create-route, routes, unmap-route"`
}

func (cmd MapRouteCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
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
		if _, ok := err.(actionerror.RouteNotFoundError); !ok {
			return err
		}
		cmd.UI.DisplayTextWithFlavor("Creating route {{.URL}} for org {{.OrgName}} / space {{.SpaceName}} as {{.User}}...",
			map[string]interface{}{
				"URL":       url,
				"User":      user.Name,
				"SpaceName": cmd.Config.TargetedSpace().Name,
				"OrgName":   cmd.Config.TargetedOrganization().Name,
			})
		route, warnings, err = cmd.Actor.CreateRoute(
			cmd.Config.TargetedSpace().GUID,
			domain.Name,
			cmd.Hostname,
			path,
			cmd.Port,
		)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}
		cmd.UI.DisplayOK()
	}

	cmd.UI.DisplayTextWithFlavor("Mapping route {{.URL}} to app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.User}}...", map[string]interface{}{
		"URL":       route.URL,
		"AppName":   cmd.RequiredArgs.App,
		"User":      user.Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
	})
	dest, err := cmd.Actor.GetRouteDestinationByAppGUID(route, app.GUID)
	if err != nil {
		if _, ok := err.(actionerror.RouteDestinationNotFoundError); !ok {
			return err
		}
	}
	if dest.GUID != "" {
		cmd.UI.DisplayText("App '{{ .AppName }}' is already mapped to route '{{ .URL}}'.", map[string]interface{}{
			"AppName": cmd.RequiredArgs.App,
			"URL":     route.URL,
		})
		cmd.UI.DisplayOK()
		return nil
	}
	warnings, err = cmd.Actor.MapRoute(route.GUID, app.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	cmd.UI.DisplayOK()

	return nil
}
