package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type MapRouteCommand struct {
	RequiredArgs    flag.AppDomain `positional-args:"yes"`
	Hostname        string         `long:"hostname" short:"n" description:"Hostname for the HTTP route (required for shared domains)"`
	Path            string         `long:"path" description:"Path for the HTTP route"`
	Port            int            `long:"port" description:"Port for the TCP route"`
	RandomPort      bool           `long:"random-port" description:"Create a random port for the TCP route"`
	usage           interface{}    `usage:"Map an HTTP route:\n      CF_NAME map-route APP_NAME DOMAIN [--hostname HOSTNAME] [--path PATH]\n\n   Map a TCP route:\n      CF_NAME map-route APP_NAME DOMAIN (--port PORT | --random-port)\n\nEXAMPLES:\n   CF_NAME map-route my-app example.com                              # example.com\n   CF_NAME map-route my-app example.com --hostname myhost            # myhost.example.com\n   CF_NAME map-route my-app example.com --hostname myhost --path foo # myhost.example.com/foo\n   CF_NAME map-route my-app example.com --port 5000                  # example.com:5000"`
	relatedCommands interface{}    `related_commands:"create-route, routes"`
}

func (MapRouteCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (MapRouteCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
