package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type AddPluginRepoCommand struct{}

func (_ AddPluginRepoCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
