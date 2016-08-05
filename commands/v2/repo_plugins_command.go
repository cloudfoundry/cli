package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type RepoPluginsCommand struct {
	RegisteredRepository string `short:"r" description:"Name of a registered repository"`
}

func (_ RepoPluginsCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
