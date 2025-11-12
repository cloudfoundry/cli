package v7

import (
	"fmt"

	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/v9/command"
	"code.cloudfoundry.org/cli/v9/command/flag"
	"code.cloudfoundry.org/cli/v9/resources"
)

type UpdateRouteCommand struct {
	BaseCommand

	RequiredArgs    flag.Domain      `positional-args:"yes"`
	Hostname        string           `long:"hostname" short:"n" description:"Hostname for the HTTP route (required for shared domains)"`
	Path            flag.V7RoutePath `long:"path" description:"Path for the HTTP route"`
	Options         []string         `long:"option" short:"o" description:"Set the value of a per-route option"`
	RemoveOptions   []string         `long:"remove-option" short:"r" description:"Remove an option with the given name"`
	relatedCommands interface{}      `related_commands:"check-route, domains, map-route, routes, unmap-route"`
}

func (cmd UpdateRouteCommand) Usage() string {
	return `
Update an existing HTTP route:
   CF_NAME update-route DOMAIN [--hostname HOSTNAME] [--path PATH] [--option OPTION=VALUE] [--remove-option OPTION]`
}

func (cmd UpdateRouteCommand) Examples() string {
	return `
CF_NAME update-route example.com -o loadbalancing=round-robin,
CF_NAME update-route example.com -o loadbalancing=least-connection,
CF_NAME update-route example.com -r loadbalancing,
CF_NAME update-route example.com --hostname myhost --path foo -o loadbalancing=round-robin`
}
func (cmd UpdateRouteCommand) Execute(args []string) error {
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

	path := cmd.Path.Path

	route, warnings, err := cmd.Actor.GetRouteByAttributes(domain, cmd.Hostname, path, 0)
	url := desiredURL(domain.Name, cmd.Hostname, path, 0)
	cmd.UI.DisplayWarnings(warnings)

	if err != nil {
		if _, ok := err.(actionerror.RouteNotFoundError); !ok {
			return err
		}
	}

	// Update route only works for per route options. The command will fail instead instead of just
	// ignoring the per-route options like it is the case for create-route and map-route.
	err = cmd.validateAPIVersionForPerRouteOptions()
	if err != nil {
		return err
	}

	if cmd.Options == nil && cmd.RemoveOptions == nil {
		return actionerror.RouteOptionSupportError{
			ErrorText: fmt.Sprintf("No options were specified for the update of the Route %s", route.URL)}
	}

	if len(cmd.Options) > 0 {
		routeOpts, wrongOptSpec := resources.CreateRouteOptions(cmd.Options)
		if wrongOptSpec != nil {
			return actionerror.RouteOptionError{
				Name:       *wrongOptSpec,
				DomainName: domain.Name,
				Path:       path,
				Host:       cmd.Hostname,
			}
		}

		cmd.UI.DisplayTextWithFlavor("Updating route {{.URL}} for org {{.OrgName}} / space {{.SpaceName}} as {{.User}}...",
			map[string]interface{}{
				"URL":       url,
				"User":      user.Name,
				"SpaceName": cmd.Config.TargetedSpace().Name,
				"OrgName":   cmd.Config.TargetedOrganization().Name,
			})
		route, warnings, err = cmd.Actor.UpdateRoute(
			route.GUID,
			routeOpts,
		)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}
	}

	if cmd.RemoveOptions != nil {
		inputRouteOptions := resources.RemoveRouteOptions(cmd.RemoveOptions)
		route, warnings, err = cmd.Actor.UpdateRoute(
			route.GUID,
			inputRouteOptions,
		)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}
	}

	cmd.UI.DisplayText("Route {{.URL}} has been updated",
		map[string]interface{}{
			"URL": route.URL,
		})
	cmd.UI.DisplayOK()

	return nil
}

func (cmd UpdateRouteCommand) validateAPIVersionForPerRouteOptions() error {
	err := command.MinimumCCAPIVersionCheck(cmd.Config.APIVersion(), ccversion.MinVersionPerRouteOpts)
	if err != nil {
		cmd.UI.DisplayWarning("Your CC API version ({{.APIVersion}}) does not support per-route options."+
			"Upgrade to a newer version of the API (minimum version {{.MinSupportedVersion}}). ", map[string]interface{}{
			"APIVersion":          cmd.Config.APIVersion(),
			"MinSupportedVersion": ccversion.MinVersionPerRouteOpts,
		})
	}
	return err
}
