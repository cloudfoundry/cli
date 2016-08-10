package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type ListPluginReposCommand struct {
	usage interface{} `usage:"CF_NAME list-plugin-repos"`
}

func (_ ListPluginReposCommand) Setup() error {
	return nil
}

func (_ ListPluginReposCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
