package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type UpdateSecurityGroupCommand struct {
	RequiredArgs flags.SecurityGroupArgs `positional-args:"yes"`
}

func (_ UpdateSecurityGroupCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
