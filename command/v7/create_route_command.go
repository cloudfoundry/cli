package v7

import (
	"fmt"

	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/v9/command"
	"code.cloudfoundry.org/cli/v9/resources"

    "code.cloudfoundry.org/cli/v9/actor/actionerror"
    "code.cloudfoundry.org/cli/v9/command/flag"
)

type CreateRouteCommand struct {
	BaseCommand

	RequiredArgs    flag.Domain      `positional-args:"yes"`
	Hostname        string           `long:"hostname" short:"n" description:"Hostname for the HTTP route (required for shared domains)"`
	Path            flag.V7RoutePath `long:"path" description:"Path for the HTTP route"`
	Port            int              `long:"port" description:"Port for the TCP route (default: random port)"`
	Options         []string         `long:"option" short:"o" description:"Set the value of a per-route option"`
	relatedCommands interface{}      `related_commands:"check-route, update-route, domains, map-route, routes, unmap-route"`
}

func (cmd CreateRouteCommand) Usage() string {
	return `
Create an HTTP route:
   CF_NAME create-route DOMAIN [--hostname HOSTNAME] [--path PATH] [--option OPTION=VALUE]
Create a TCP route:
   CF_NAME create-route DOMAIN [--port PORT] [--option OPTION=VALUE]`
}

func (cmd CreateRouteCommand) Examples() string {
	return `
CF_NAME create-route example.com                                                     # example.com
CF_NAME create-route example.com --hostname myapp                                    # myapp.example.com
CF_NAME create-route example.com --hostname myapp --path foo                         # myapp.example.com/foo
CF_NAME create-route example.com --port 5000                                         # example.com:5000
CF_NAME create-route example.com --hostname myapp -o loadbalancing=least-connection  # myapp.example.com with a per-route option
`
}

func (cmd CreateRouteCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	domain := cmd.RequiredArgs.Domain
	hostname := cmd.Hostname
	pathName := cmd.Path.Path
	port := cmd.Port
	spaceName := cmd.Config.TargetedSpace().Name
	orgName := cmd.Config.TargetedOrganization().Name
	spaceGUID := cmd.Config.TargetedSpace().GUID
	url := desiredURL(domain, hostname, pathName, port)

	cmd.UI.DisplayTextWithFlavor("Creating route {{.URL}} for org {{.Organization}} / space {{.Space}} as {{.User}}...",
		map[string]interface{}{
			"URL":          url,
			"User":         user.Name,
			"Space":        spaceName,
			"Organization": orgName,
		})

	var routeOptions map[string]*string
	if len(cmd.Options) > 0 && cmd.validateAPIVersionForPerRouteOptions() == nil {
		var wrongOptSpec *string
		routeOptions, wrongOptSpec = resources.CreateRouteOptions(cmd.Options)
		if wrongOptSpec != nil {
			return actionerror.RouteOptionError{
				Name:       *wrongOptSpec,
				DomainName: domain,
				Path:       pathName,
				Host:       hostname,
			}
		}
	}
	route, warnings, err := cmd.Actor.CreateRoute(spaceGUID, domain, hostname, pathName, port, routeOptions)

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		if _, ok := err.(actionerror.RouteAlreadyExistsError); ok {
			cmd.UI.DisplayWarning(err.Error())
			cmd.UI.DisplayOK()
			return nil
		}
		return err
	}

	cmd.UI.DisplayText("Route {{.URL}} has been created.",
		map[string]interface{}{
			"URL": route.URL,
		})

	cmd.UI.DisplayOK()
	return nil
}

func desiredURL(domain, hostname, path string, port int) string {
	url := ""

	if hostname != "" {
		url += hostname + "."
	}

	url += domain

	if path != "" {
		url += path
	}

	if port != 0 {
		url += fmt.Sprintf(":%d", port)
	}

	return url
}

func (cmd CreateRouteCommand) validateAPIVersionForPerRouteOptions() error {
	err := command.MinimumCCAPIVersionCheck(cmd.Config.APIVersion(), ccversion.MinVersionPerRouteOpts)
	if err != nil {
		cmd.UI.DisplayWarning("Your CC API version ({{.APIVersion}}) does not support per-route options. Those will be ignored. Upgrade to a newer version of the API (minimum version {{.MinSupportedVersion}}).", map[string]interface{}{
			"APIVersion":          cmd.Config.APIVersion(),
			"MinSupportedVersion": ccversion.MinVersionPerRouteOpts,
		})
	}
	return err
}
