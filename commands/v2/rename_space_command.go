package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type RenameSpaceCommand struct {
	RequiredArgs flags.RenameSpaceArgs `positional-args:"yes"`
	usage        interface{}           `usage:"CF_NAME rename-space SPACE NEW_SPACE"`
}

func (_ RenameSpaceCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ RenameSpaceCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
