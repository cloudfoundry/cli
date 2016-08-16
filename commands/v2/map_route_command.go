package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type MapRouteCommand struct {
	RequiredArgs flags.AppDomain `positional-args:"yes"`
	Hostname     string          `long:"hostname" short:"n" description:"Hostname for the HTTP route (required for shared domains)"`
	Path         string          `long:"path" description:"Path for the HTTP route"`
	Port         int             `long:"port" description:"Port for the TCP route"`
	RandomPort   bool            `long:"random-port" description:"Create a random port for the TCP route"`
	usage        interface{}     `usage:"Unmap an HTTP route:\n       CF_NAME unmap-route APP_NAME DOMAIN [--hostname HOSTNAME] [--path PATH]\n\nUnmap a TCP route:\n       CF_NAME unmap-route APP_NAME DOMAIN --port PORT\n\nEXAMPLES:\n    CF_NAME unmap-route my-app example.com                              # example.com\n    CF_NAME unmap-route my-app example.com --hostname myhost            # myhost.example.com\n    CF_NAME unmap-route my-app example.com --hostname myhost --path foo # myhost.example.com/foo\n    CF_NAME unmap-route my-app example.com --port 5000                  # example.com:5000"`
}

func (_ MapRouteCommand) Setup(config commands.Config) error {
	return nil
}

func (_ MapRouteCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
