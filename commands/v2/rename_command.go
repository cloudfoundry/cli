package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type RenameCommand struct {
	RequiredArgs flags.AppRenameArgs `positional-args:"yes"`
	usage        interface{}         `usage:"CF_NAME rename APP_NAME NEW_APP_NAME"`
}

func (_ RenameCommand) Setup() error {
	return nil
}

func (_ RenameCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
