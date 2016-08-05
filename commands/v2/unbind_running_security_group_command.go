package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type UnbindRunningSecurityGroupCommand struct {
	RequiredArgs flags.SecurityGroup `positional-args:"yes"`
}

func (_ UnbindRunningSecurityGroupCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
