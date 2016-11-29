package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type UnbindRouteServiceCommand struct {
	RequiredArgs    flag.RouteServiceArgs `positional-args:"yes"`
	Force           bool                  `short:"f" description:"Force unbinding without confirmation"`
	Hostname        string                `long:"hostname" short:"n" description:"Hostname used in combination with DOMAIN to specify the route to unbind"`
	Path            string                `long:"path" description:"Path used in combination with HOSTNAME and DOMAIN to specify the route to unbind"`
	usage           interface{}           `usage:"CF_NAME unbind-route-service DOMAIN SERVICE_INSTANCE [--hostname HOSTNAME] [--path PATH] [-f]\n\nEXAMPLES:\n   CF_NAME unbind-route-service example.com myratelimiter --hostname myapp --path foo"`
	relatedCommands interface{}           `related_commands:"delete-service, routes, services"`
}

func (_ UnbindRouteServiceCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ UnbindRouteServiceCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
