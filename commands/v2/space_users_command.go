package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type SpaceUsersCommand struct {
	RequiredArgs flags.OrgSpace `positional-args:"yes"`
}

func (_ SpaceUsersCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
