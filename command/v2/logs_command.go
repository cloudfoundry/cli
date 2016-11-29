package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type LogsCommand struct {
	RequiredArgs    flag.AppName `positional-args:"yes"`
	Recent          bool         `long:"recent" description:"Dump recent logs instead of tailing"`
	usage           interface{}  `usage:"CF_NAME logs APP_NAME"`
	relatedCommands interface{}  `related_commands:"app, apps, ssh"`
}

func (_ LogsCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ LogsCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
