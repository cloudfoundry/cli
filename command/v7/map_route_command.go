package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
)

//go:generate counterfeiter . MapRouteActor

type MapRouteActor interface {
	GetApplicationsByNamesAndSpace(appNames []string, spaceGUID string) ([]v7action.Application, v7action.Warnings, error)
	GetRouteByAttributesAndSpace(domainGUID string, hostname string, path string, spaceGUID string) (v7action.Route, v7action.Warnings, error)
	GetDomainByName(domainName string) (v7action.Domain, v7action.Warnings, error)
	CreateRoute(orgName, spaceName, domainName, hostname, path string) (v7action.Route, v7action.Warnings, error)
	MapRoute(routeGUID string, appGUID string) (v7action.Warnings, error)
}

type MapRouteCommand struct {
	RequiredArgs    flag.AppDomain `positional-args:"yes"`
	usage           interface{}    `usage:"CF_NAME map-route APP_NAME DOMAIN [--hostname HOSTNAME] [--path PATH]\n\nEXAMPLES:\n   CF_NAME map-route my-app example.com                              # example.com\n   CF_NAME map-route my-app example.com --hostname myhost            # myhost.example.com\n   CF_NAME map-route my-app example.com --hostname myhost --path foo # myhost.example.com/foo"`
	Hostname        string         `long:"hostname" short:"n" description:"Hostname for the HTTP route (required for shared domains)"`
	Path            string         `long:"path" description:"Path for the HTTP route"`
	relatedCommands interface{}    `related_commands:"create-route, routes, unmap-route"`

	UI          command.UI
	Config      command.Config
	Actor       MapRouteActor
	SharedActor command.SharedActor
}

func (cmd *MapRouteCommand) Setup(config command.Config, ui command.UI) error {
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
	apps, warnings, err := cmd.Actor.GetApplicationsByNamesAndSpace([]string{cmd.RequiredArgs.App}, spaceGUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	route, warnings, err := cmd.Actor.GetRouteByAttributesAndSpace(domain.GUID, cmd.Hostname, cmd.Path, spaceGUID)
	fqdn := desiredFQDN(domain.Name, cmd.Hostname, cmd.Path)
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
			cmd.Config.TargetedOrganization().Name,
			cmd.Config.TargetedSpace().Name,
			domain.Name,
			cmd.Hostname,
			cmd.Path,
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
	warnings, err = cmd.Actor.MapRoute(route.GUID, apps[0].GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	cmd.UI.DisplayOK()

	return nil
}

