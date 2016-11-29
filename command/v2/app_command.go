package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type AppCommand struct {
	RequiredArgs    flag.AppName `positional-args:"yes"`
	GUID            bool         `long:"guid" description:"Retrieve and display the given app's guid.  All other health and status output for the app is suppressed."`
	usage           interface{}  `usage:"CF_NAME app APP_NAME"`
	relatedCommands interface{}  `related_commands:"apps, events, logs, map-route, unmap-route, push"`
}

func (_ AppCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ AppCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
