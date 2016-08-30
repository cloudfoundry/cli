package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type StopCommand struct {
	RequiredArgs    flags.AppName `positional-args:"yes"`
	usage           interface{}   `usage:"CF_NAME stop APP_NAME"`
	relatedCommands interface{}   `related_commands:"restart, scale, start"`
}

func (_ StopCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ StopCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
