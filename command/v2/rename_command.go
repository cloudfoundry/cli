package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flags"
)

type RenameCommand struct {
	RequiredArgs    flags.AppRenameArgs `positional-args:"yes"`
	usage           interface{}         `usage:"CF_NAME rename APP_NAME NEW_APP_NAME"`
	relatedCommands interface{}         `related_commands:"apps, delete"`
}

func (_ RenameCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ RenameCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
