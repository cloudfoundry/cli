package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type AddPluginRepoCommand struct {
	RequiredArgs flags.AddPluginRepoArgs `positional-args:"yes"`
}

func (_ AddPluginRepoCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
