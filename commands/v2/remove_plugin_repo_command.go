package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type RemovePluginRepoCommand struct {
	RequiredArgs flags.PluginRepoName `positional-args:"yes"`
}

func (_ RemovePluginRepoCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
