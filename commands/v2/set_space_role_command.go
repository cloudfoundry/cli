package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type SetSpaceRoleCommand struct {
	RequiredArgs flags.SetSpaceRoleArgs `positional-args:"yes"`
}

func (_ SetSpaceRoleCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
