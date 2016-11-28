package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
)

type ListPluginReposCommand struct {
	usage           interface{} `usage:"CF_NAME list-plugin-repos"`
	relatedCommands interface{} `related_commands:"add-plugin-repo, install-plugin"`
}

func (_ ListPluginReposCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ ListPluginReposCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
