package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type BindStagingSecurityGroupCommand struct {
	RequiredArgs flags.SecurityGroup `positional-args:"yes"`
}

func (_ BindStagingSecurityGroupCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
