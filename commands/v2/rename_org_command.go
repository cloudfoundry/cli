package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type RenameOrgCommand struct {
	RequiredArgs flags.RenameOrgArgs `positional-args:"yes"`
	usage        interface{}         `usage:"CF_NAME rename-org ORG NEW_ORG"`
}

func (_ RenameOrgCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ RenameOrgCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
