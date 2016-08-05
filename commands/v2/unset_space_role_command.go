package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type UnsetSpaceRoleCommand struct {
	RequiredArgs flags.SetSpaceRoleArgs `positional-args:"yes"`
}

func (_ UnsetSpaceRoleCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
