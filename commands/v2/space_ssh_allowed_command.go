package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type SpaceSSHAllowedCommand struct {
	RequiredArgs flags.Space `positional-args:"yes"`
}

func (_ SpaceSSHAllowedCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
