package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type UnmapRouteCommand struct {
	RequiredArgs    flag.AppDomain `positional-args:"yes"`
	Hostname        string         `long:"hostname" short:"n" description:"Hostname used to identify the HTTP route"`
	Path            string         `long:"path" description:"Path used to identify the HTTP route"`
	Port            int            `long:"port" description:"Port used to identify the TCP route"`
	usage           interface{}    `usage:"Unmap an HTTP route:\n      CF_NAME unmap-route APP_NAME DOMAIN [--hostname HOSTNAME] [--path PATH]\n\n   Unmap a TCP route:\n      CF_NAME unmap-route APP_NAME DOMAIN --port PORT\n\nEXAMPLES:\n   CF_NAME unmap-route my-app example.com                              # example.com\n   CF_NAME unmap-route my-app example.com --hostname myhost            # myhost.example.com\n   CF_NAME unmap-route my-app example.com --hostname myhost --path foo # myhost.example.com/foo\n   CF_NAME unmap-route my-app example.com --port 5000                  # example.com:5000"`
	relatedCommands interface{}    `related_commands:"delete-route, routes"`
}

func (_ UnmapRouteCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ UnmapRouteCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
