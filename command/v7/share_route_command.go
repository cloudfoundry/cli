package v7

import (
	"code.cloudfoundry.org/cli/v8/actor/actionerror"
	"code.cloudfoundry.org/cli/v8/command/flag"
)

type ShareRouteCommand struct {
	BaseCommand

	RequireArgs      flag.Domain      `positional-args:"yes"`
	Hostname         string           `long:"hostname" short:"n" description:"Hostname for the HTTP route (required for shared domains)"`
	Path             flag.V7RoutePath `long:"path" description:"Path for the HTTP route"`
	DestinationOrg   string           `short:"o" description:"The org of the destination space (Default: targeted org)"`
	DestinationSpace string           `short:"s" description:"The space the route will be shared with (Default: targeted space)"`

	relatedCommands interface{} `related_commands:"create-route, map-route, unmap-route, routes"`
}

func (cmd ShareRouteCommand) Usage() string {
	return `
Share an existing route in between two spaces:
	CF_NAME share-route DOMAIN [--hostname HOSTNAME] [--path PATH] -s OTHER_SPACE [-o OTHER_ORG]`
}

func (cmd ShareRouteCommand) Examples() string {
	return `
CF_NAME share-route example.com --hostname myHost --path foo -s TargetSpace -o TargetOrg        # myhost.example.com/foo
CF_NAME share-route example.com --hostname myHost -s TargetSpace                                # myhost.example.com
CF_NAME share-route example.com --hostname myHost -s TargetSpace -o TargetOrg                   # myhost.example.com`
}

func (cmd ShareRouteCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	domain, warnings, err := cmd.Actor.GetDomainByName(cmd.RequireArgs.Domain)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	path := cmd.Path.Path
	route, warnings, err := cmd.Actor.GetRouteByAttributes(domain, cmd.Hostname, path, 0)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		if _, ok := err.(actionerror.RouteNotFoundError); ok {
			cmd.UI.DisplayText("Can not share route:")
			return err
		}
	}

	destinationOrgName := cmd.DestinationOrg

	if destinationOrgName == "" {
		destinationOrgName = cmd.Config.TargetedOrganizationName()
	}

	destinationOrg, warnings, err := cmd.Actor.GetOrganizationByName(destinationOrgName)
	cmd.UI.DisplayWarnings(warnings)

	if err != nil {
		if _, ok := err.(actionerror.OrganizationNotFoundError); ok {
			cmd.UI.DisplayText("Can not share route:")
			return err
		}
	}

	targetedSpace, warnings, err := cmd.Actor.GetSpaceByNameAndOrganization(cmd.DestinationSpace, destinationOrg.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		if _, ok := err.(actionerror.SpaceNotFoundError); ok {
			cmd.UI.DisplayText("Can not share route:")
			return err
		}
	}

	url := desiredURL(domain.Name, cmd.Hostname, path, 0)
	cmd.UI.DisplayTextWithFlavor("Sharing route {{.URL}} to space {{.DestinationSpace}} as {{.User}}",
		map[string]interface{}{
			"URL":              url,
			"DestinationSpace": cmd.DestinationSpace,
			"User":             user.Name,
		})
	warnings, err = cmd.Actor.ShareRoute(
		route.GUID,
		targetedSpace.GUID,
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	cmd.UI.DisplayOK()

	return nil
}
