package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type RenameOrgCommand struct {
	RequiredArgs flag.RenameOrgArgs `positional-args:"yes"`
	usage        interface{}        `usage:"CF_NAME rename-org ORG NEW_ORG"`
}

func (_ RenameOrgCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ RenameOrgCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
