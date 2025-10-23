package v7

import (
	"strconv"

	"code.cloudfoundry.org/cli/v8/command/flag"
	"code.cloudfoundry.org/cli/v8/resources"
)

type RouteCommand struct {
	BaseCommand

	RequiredArgs    flag.Domain      `positional-args:"yes"`
	Hostname        string           `long:"hostname" short:"n" description:"Hostname used to identify the HTTP route"`
	Path            flag.V7RoutePath `long:"path" description:"Path used to identify the HTTP route"`
	Port            int              `long:"port" description:"Port used to identify the TCP route"`
	relatedCommands interface{}      `related_commands:"create-route, delete-route, routes"`
}

func (cmd RouteCommand) Usage() string {
	return `
Display an HTTP route:
   CF_NAME route DOMAIN [--hostname HOSTNAME] [--path PATH]

Display a TCP route:
   CF_NAME route DOMAIN --port PORT`
}

func (cmd RouteCommand) Examples() string {
	return `
CF_NAME route example.com                      # example.com
CF_NAME route example.com -n myhost --path foo # myhost.example.com/foo
CF_NAME route example.com --path foo           # example.com/foo
CF_NAME route example.com --port 5000          # example.com:5000`
}

func (cmd RouteCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, false)
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

	hostName := ""
	if cmd.Hostname != "" {
		hostName = cmd.Hostname + "."
	}

	displayPort := ""
	if cmd.Port != 0 {
		displayPort = ":" + strconv.Itoa(cmd.Port)

	}

	cmd.UI.DisplayTextWithFlavor(" Showing route {{.HostName}}{{.DomainName}}{{.Port}}{{.PathName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"HostName":   hostName,
		"DomainName": cmd.RequiredArgs.Domain,
		"PathName":   cmd.Path.Path,
		"Port":       displayPort,
		"OrgName":    cmd.Config.TargetedOrganization().Name,
		"SpaceName":  cmd.Config.TargetedSpace().Name,
		"Username":   user.Name,
	})
	cmd.UI.DisplayNewline()

	route, warnings, err := cmd.Actor.GetRouteByAttributes(domain, cmd.Hostname, cmd.Path.Path, cmd.Port)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	port := ""
	if route.Port != 0 {
		port = strconv.Itoa(route.Port)
	}

	appMap, warnings, err := cmd.Actor.GetApplicationMapForRoute(route)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	table := [][]string{
		{cmd.UI.TranslateText("domain:"), domain.Name},
		{cmd.UI.TranslateText("host:"), route.Host},
		{cmd.UI.TranslateText("port:"), port},
		{cmd.UI.TranslateText("path:"), route.Path},
		{cmd.UI.TranslateText("protocol:"), route.Protocol},
		{cmd.UI.TranslateText("options:"), route.FormattedOptions()},
	}

	cmd.UI.DisplayKeyValueTable("", table, 3)
	cmd.UI.DisplayNewline()

	cmd.UI.DisplayText("Destinations:")
	cmd.displayDestinations(route, appMap)

	return nil
}

func (cmd RouteCommand) displayDestinations(route resources.Route, appMap map[string]resources.Application) {
	destinations := route.Destinations
	if len(destinations) > 0 {
		var keyValueTable = [][]string{
			{
				cmd.UI.TranslateText("app"),
				cmd.UI.TranslateText("process"),
				cmd.UI.TranslateText("port"),
				cmd.UI.TranslateText("app-protocol"),
			},
		}

		for _, destination := range destinations {
			port := ""
			if destination.Port != 0 {
				port = strconv.Itoa(destination.Port)
			}
			keyValueTable = append(keyValueTable, []string{
				appMap[destination.App.GUID].Name,
				destination.App.Process.Type,
				port,
				destination.Protocol,
			})
		}

		cmd.UI.DisplayKeyValueTable("\t", keyValueTable, 3)
	}
}
