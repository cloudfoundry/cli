package v7

import (
	"code.cloudfoundry.org/cli/v8/actor/actionerror"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/v8/command"
	"code.cloudfoundry.org/cli/v8/command/flag"
	"code.cloudfoundry.org/cli/v8/resources"
)

type MapRouteCommand struct {
	BaseCommand

	RequiredArgs    flag.AppDomain   `positional-args:"yes"`
	Hostname        string           `long:"hostname" short:"n" description:"Hostname for the HTTP route (required for shared domains)"`
	Path            flag.V7RoutePath `long:"path" description:"Path for the HTTP route"`
	Port            int              `long:"port" description:"Port for the TCP route (default: random port)"`
	AppProtocol     string           `long:"app-protocol" description:"[Beta flag, subject to change] Protocol for the route destination (default: http1). Only applied to HTTP routes"`
	Options         []string         `long:"option" short:"o" description:"Set the value of a per-route option"`
	relatedCommands interface{}      `related_commands:"create-route, update-route, routes, unmap-route"`
}

func (cmd MapRouteCommand) Usage() string {
	return `
Map an HTTP route:
   CF_NAME map-route APP_NAME DOMAIN [--hostname HOSTNAME] [--path PATH] [--app-protocol PROTOCOL] [--option OPTION=VALUE]

Map a TCP route:
   CF_NAME map-route APP_NAME DOMAIN [--port PORT] [--option OPTION=VALUE]`
}

func (cmd MapRouteCommand) Examples() string {
	return `
CF_NAME map-route my-app example.com                                                      # example.com
CF_NAME map-route my-app example.com --hostname myhost                                    # myhost.example.com
CF_NAME map-route my-app example.com --hostname myhost -o loadbalancing=least-connection  # myhost.example.com with a per-route option
CF_NAME map-route my-app example.com --hostname myhost --path foo                         # myhost.example.com/foo
CF_NAME map-route my-app example.com --hostname myhost --app-protocol http2               # myhost.example.com
CF_NAME map-route my-app example.com --port 5000                                          # example.com:5000`
}

func (cmd MapRouteCommand) Execute(args []string) error {
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
		if _, ok := err.(actionerror.RouteNotFoundError); !ok {
			return err
		}

		var routeOptions map[string]*string
		if len(cmd.Options) > 0 && cmd.validateAPIVersionForPerRouteOptions() == nil {
			var wrongOptSpec *string
			routeOptions, wrongOptSpec = resources.CreateRouteOptions(cmd.Options)
			if wrongOptSpec != nil {
				return actionerror.RouteOptionError{
					Name:       *wrongOptSpec,
					DomainName: domain.Name,
					Path:       path,
					Host:       cmd.Hostname,
				}
			}
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
			routeOptions,
		)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}
		cmd.UI.DisplayOK()
	} else {
		if len(cmd.Options) > 0 {
			return actionerror.RouteOptionSupportError{ErrorText: "Route specific options can only be specified for nonexistent routes."}
		}
	}

	if cmd.AppProtocol != "" {
		cmd.UI.DisplayTextWithFlavor("Mapping route {{.URL}} to app {{.AppName}} with protocol {{.Protocol}} in org {{.OrgName}} / space {{.SpaceName}} as {{.User}}...", map[string]interface{}{
			"URL":       route.URL,
			"AppName":   cmd.RequiredArgs.App,
			"User":      user.Name,
			"SpaceName": cmd.Config.TargetedSpace().Name,
			"OrgName":   cmd.Config.TargetedOrganization().Name,
			"Protocol":  cmd.AppProtocol,
		})

	} else {
		cmd.UI.DisplayTextWithFlavor("Mapping route {{.URL}} to app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.User}}...", map[string]interface{}{
			"URL":       route.URL,
			"AppName":   cmd.RequiredArgs.App,
			"User":      user.Name,
			"SpaceName": cmd.Config.TargetedSpace().Name,
			"OrgName":   cmd.Config.TargetedOrganization().Name,
		})
	}
	dest, err := cmd.Actor.GetRouteDestinationByAppGUID(route, app.GUID)
	if err != nil {
		if _, ok := err.(actionerror.RouteDestinationNotFoundError); !ok {
			return err
		}
	}
	if dest.GUID != "" {
		cmd.UI.DisplayText("App '{{ .AppName }}' is already mapped to route '{{ .URL}}'. Nothing has been updated.", map[string]interface{}{
			"AppName": cmd.RequiredArgs.App,
			"URL":     route.URL,
		})
		cmd.UI.DisplayOK()
		return nil
	}

	warnings, err = cmd.Actor.MapRoute(route.GUID, app.GUID, cmd.AppProtocol)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	cmd.UI.DisplayOK()

	return nil
}

func (cmd MapRouteCommand) validateAPIVersionForPerRouteOptions() error {
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
