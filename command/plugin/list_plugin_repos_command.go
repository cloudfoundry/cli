package plugin

import (
	"code.cloudfoundry.org/cli/v9/command"
	"code.cloudfoundry.org/cli/v9/command/translatableerror"
)

type ListPluginReposCommand struct {
	usage           interface{} `usage:"CF_NAME list-plugin-repos"`
	relatedCommands interface{} `related_commands:"add-plugin-repo, install-plugin"`
}

func (ListPluginReposCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (ListPluginReposCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
