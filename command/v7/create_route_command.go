package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/command/flag"
)

type CreateRouteCommand struct {
	command.BaseCommand

	RequiredArgs    flag.Domain      `positional-args:"yes"`
	usage           interface{}      `usage:"CF_NAME create-route DOMAIN [--hostname HOSTNAME] [--path PATH]\n\nEXAMPLES:\n   CF_NAME create-route example.com                             # example.com\n   CF_NAME create-route example.com --hostname myapp            # myapp.example.com\n   CF_NAME create-route example.com --hostname myapp --path foo # myapp.example.com/foo"`
	Hostname        string           `long:"hostname" short:"n" description:"Hostname for the HTTP route (required for shared domains)"`
	Path            flag.V7RoutePath `long:"path" description:"Path for the HTTP route"`
	relatedCommands interface{}      `related_commands:"check-route, domains, map-route, routes, unmap-route"`
}

func (cmd CreateRouteCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	domain := cmd.RequiredArgs.Domain
	hostname := cmd.Hostname
	pathName := cmd.Path.Path
	spaceName := cmd.Config.TargetedSpace().Name
	orgName := cmd.Config.TargetedOrganization().Name
	spaceGUID := cmd.Config.TargetedSpace().GUID
	fqdn := desiredFQDN(domain, hostname, pathName)

	cmd.UI.DisplayTextWithFlavor("Creating route {{.FQDN}} for org {{.Organization}} / space {{.Space}} as {{.User}}...",
		map[string]interface{}{
			"FQDN":         fqdn,
			"User":         user.Name,
			"Space":        spaceName,
			"Organization": orgName,
		})

	_, warnings, err := cmd.Actor.CreateRoute(spaceGUID, domain, hostname, pathName)

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		if _, ok := err.(actionerror.RouteAlreadyExistsError); ok {
			cmd.UI.DisplayWarning(err.Error())
			cmd.UI.DisplayOK()
			return nil
		}
		return err
	}

	cmd.UI.DisplayText("Route {{.FQDN}} has been created.",
		map[string]interface{}{
			"FQDN": fqdn,
		})

	cmd.UI.DisplayOK()
	return nil
}

func desiredFQDN(domain, hostname, path string) string {
	fqdn := ""

	if hostname != "" {
		fqdn += hostname + "."
	}
	fqdn += domain

	if path != "" {
		fqdn += path
	}

	return fqdn
}
