package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . UnmapRouteActor

type UnmapRouteActor interface {
	GetApplicationByNameAndSpace(appName string, spaceGUID string) (v7action.Application, v7action.Warnings, error)
	GetRouteByAttributes(domainName string, domainGUID string, hostname string, path string) (v7action.Route, v7action.Warnings, error)
	GetDomainByName(domainName string) (resources.Domain, v7action.Warnings, error)
	GetRouteDestinationByAppGUID(routeGUID string, appGUID string) (v7action.RouteDestination, v7action.Warnings, error)
	UnmapRoute(routeGUID string, destinationGUID string) (v7action.Warnings, error)
}

type UnmapRouteCommand struct {
	RequiredArgs    flag.AppDomain   `positional-args:"yes"`
	Hostname        string           `long:"hostname" short:"n" description:"Hostname used to identify the HTTP route"`
	Path            flag.V7RoutePath `long:"path" description:"Path used to identify the HTTP route"`
	usage           interface{}      `usage:"CF_NAME unmap-route APP_NAME DOMAIN [--hostname HOSTNAME] [--path PATH]\n\nEXAMPLES:\n   CF_NAME unmap-route my-app example.com                              # example.com\n   CF_NAME unmap-route my-app example.com --hostname myhost            # myhost.example.com\n   CF_NAME unmap-route my-app example.com --hostname myhost --path foo # myhost.example.com/foo"`
	relatedCommands interface{}      `related_commands:"delete-route, map-route, routes"`

	UI          command.UI
	Config      command.Config
	Actor       UnmapRouteActor
	SharedActor command.SharedActor
}

func (cmd *UnmapRouteCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.GetNewClientsAndConnectToCF(config, ui, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, sharedActor, uaaClient, clock.NewClock())
	return nil
}

func (cmd UnmapRouteCommand) Execute(args []string) error {
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
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Removing route {{.FQDN}} from app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.User}}...", map[string]interface{}{
		"FQDN":      fqdn,
		"AppName":   cmd.RequiredArgs.App,
		"User":      user.Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
	})

	destination, warnings, err := cmd.Actor.GetRouteDestinationByAppGUID(route.GUID, app.GUID)
	cmd.UI.DisplayWarnings(warnings)
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
