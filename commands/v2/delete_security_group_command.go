package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type DeleteSecurityGroupCommand struct {
	RequiredArgs flags.SecurityGroup `positional-args:"yes"`
	Force        bool                `short:"f" description:"Force deletion without confirmation"`
}

func (_ DeleteSecurityGroupCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
