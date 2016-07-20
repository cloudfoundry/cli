package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type RemovePluginRepoCommand struct{}

func (_ RemovePluginRepoCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
