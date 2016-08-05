package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type RenameOrgCommand struct {
	RequiredArgs flags.RenameOrgArgs `positional-args:"yes"`
}

func (_ RenameOrgCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
