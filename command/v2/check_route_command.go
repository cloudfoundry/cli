package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type CheckRouteCommand struct {
	RequiredArgs    flag.HostDomain `positional-args:"yes"`
	Path            string          `long:"path" description:"Path for the route"`
	usage           interface{}     `usage:"CF_NAME check-route HOST DOMAIN [--path PATH]\n\nEXAMPLES:\n   CF_NAME check-route myhost example.com            # example.com\n   CF_NAME check-route myhost example.com --path foo # myhost.example.com/foo"`
	relatedCommands interface{}     `related_commands:"create-route, delete-route, routes"`
}

func (_ CheckRouteCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ CheckRouteCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
