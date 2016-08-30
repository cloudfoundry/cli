package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type RestartCommand struct {
	RequiredArgs    flags.AppName `positional-args:"yes"`
	usage           interface{}   `usage:"CF_NAME restart APP_NAME"`
	relatedCommands interface{}   `related_commands:"restage, restart-app-instance"`
}

func (_ RestartCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ RestartCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
