package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type StopCommand struct {
	RequiredArgs    flag.AppName `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME stop APP_NAME"`
	relatedCommands interface{}  `related_commands:"restart, scale, start"`
}

func (_ StopCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ StopCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
