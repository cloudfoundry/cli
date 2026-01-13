package plugin

import (
	"code.cloudfoundry.org/cli/v9/command"
	"code.cloudfoundry.org/cli/v9/command/translatableerror"
)

type RepoPluginsCommand struct {
	RegisteredRepository string      `short:"r" description:"Name of a registered repository"`
	usage                interface{} `usage:"CF_NAME repo-plugins [-r REPO_NAME]\n\nEXAMPLES:\n   CF_NAME repo-plugins -r PrivateRepo"`
	relatedCommands      interface{} `related_commands:"add-plugin-repo, remove-plugin-repo, install-plugin"`
}

func (RepoPluginsCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (RepoPluginsCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
