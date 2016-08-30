package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type RestartAppInstanceCommand struct {
	RequiredArgs    flags.AppInstance `positional-args:"yes"`
	usage           interface{}       `usage:"CF_NAME restart-app-instance APP_NAME INDEX"`
	relatedCommands interface{}       `related_commands:"restart"`
}

func (_ RestartAppInstanceCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ RestartAppInstanceCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
