package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/flag"
)

type MapRouteActor interface {
	GetApplicationByNameAndSpace(appName string, spaceGUID string) (v7action.Application, v7action.Warnings, error)
	GetRouteByAttributes(domainName string, domainGUID string, hostname string, path string) (v7action.Route, v7action.Warnings, error)
	GetDomainByName(domainName string) (v7action.Domain, v7action.Warnings, error)
	CreateRoute(spaceGUID, domainName, hostname, path string) (v7action.Route, v7action.Warnings, error)
	GetRouteDestinationByAppGUID(routeGUID string, appGUID string) (v7action.RouteDestination, v7action.Warnings, error)
	MapRoute(routeGUID string, appGUID string) (v7action.Warnings, error)
}

type MapRouteCommand struct {
	BaseCommand

	RequiredArgs    flag.AppDomain   `positional-args:"yes"`
	usage           interface{}      `usage:"CF_NAME map-route APP_NAME DOMAIN [--hostname HOSTNAME] [--path PATH]\n\nEXAMPLES:\n   CF_NAME map-route my-app example.com                              # example.com\n   CF_NAME map-route my-app example.com --hostname myhost            # myhost.example.com\n   CF_NAME map-route my-app example.com --hostname myhost --path foo # myhost.example.com/foo"`
	Hostname        string           `long:"hostname" short:"n" description:"Hostname for the HTTP route (required for shared domains)"`
	Path            flag.V7RoutePath `long:"path" description:"Path for the HTTP route"`
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
	route, warnings, err := cmd.Actor.GetRouteByAttributes(domain.Name, domain.GUID, cmd.Hostname, path)
	fqdn := desiredFQDN(domain.Name, cmd.Hostname, path)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		if _, ok := err.(actionerror.RouteNotFoundError); !ok {
			return err
		}
		cmd.UI.DisplayTextWithFlavor("Creating route {{.FQDN}} for org {{.OrgName}} / space {{.SpaceName}} as {{.User}}...",
			map[string]interface{}{
				"FQDN":      fqdn,
				"User":      user.Name,
				"SpaceName": cmd.Config.TargetedSpace().Name,
				"OrgName":   cmd.Config.TargetedOrganization().Name,
			})
		route, warnings, err = cmd.Actor.CreateRoute(
			cmd.Config.TargetedSpace().GUID,
			domain.Name,
			cmd.Hostname,
			path,
		)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}
		cmd.UI.DisplayOK()
	}

	cmd.UI.DisplayTextWithFlavor("Mapping route {{.FQDN}} to app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.User}}...", map[string]interface{}{
		"FQDN":      fqdn,
		"AppName":   cmd.RequiredArgs.App,
		"User":      user.Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
	})
	dest, warnings, err := cmd.Actor.GetRouteDestinationByAppGUID(route.GUID, app.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		if _, ok := err.(actionerror.RouteDestinationNotFoundError); !ok {
			return err
		}
	}
	if dest.GUID != "" {
		cmd.UI.DisplayText("App '{{ .AppName }}' is already mapped to route '{{ .FQDN }}'.", map[string]interface{}{
			"AppName": cmd.RequiredArgs.App,
			"FQDN":    fqdn,
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
