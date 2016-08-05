package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type UnsetOrgRoleCommand struct {
	RequiredArgs flags.SetOrgRoleArgs `positional-args:"yes"`
}

func (_ UnsetOrgRoleCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
