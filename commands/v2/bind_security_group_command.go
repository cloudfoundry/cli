package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type BindSecurityGroupCommand struct {
	RequiredArgs flags.BindSecurityGroupArgs `positional-args:"yes"`
}

func (_ BindSecurityGroupCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
