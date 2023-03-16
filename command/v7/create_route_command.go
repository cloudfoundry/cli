package v7

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/command/flag"
)

type CreateRouteCommand struct {
	BaseCommand

	RequiredArgs    flag.Domain      `positional-args:"yes"`
	usage           interface{}      `usage:"Create an HTTP route:\n      CF_NAME create-route DOMAIN [--hostname HOSTNAME] [--path PATH]\n\n   Create a TCP route:\n      CF_NAME create-route DOMAIN [--port PORT]\n\nEXAMPLES:\n   CF_NAME create-route example.com                             # example.com\n   CF_NAME create-route example.com --hostname myapp            # myapp.example.com\n   CF_NAME create-route example.com --hostname myapp --path foo # myapp.example.com/foo\n   CF_NAME create-route example.com --port 5000                 # example.com:5000"`
	Hostname        string           `long:"hostname" short:"n" description:"Hostname for the HTTP route (required for shared domains)"`
	Path            flag.V7RoutePath `long:"path" description:"Path for the HTTP route"`
	Port            int              `long:"port" description:"Port for the TCP route (default: random port)"`
	relatedCommands interface{}      `related_commands:"check-route, domains, map-route, routes, unmap-route"`
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

	route, warnings, err := cmd.Actor.CreateRoute(spaceGUID, domain, hostname, pathName, port)

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
