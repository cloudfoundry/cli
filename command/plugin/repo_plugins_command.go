package plugin

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
)

type RepoPluginsCommand struct {
	RegisteredRepository string      `short:"r" description:"Name of a registered repository"`
	usage                interface{} `usage:"CF_NAME repo-plugins [-r REPO_NAME]\n\nEXAMPLES:\n   CF_NAME repo-plugins -r PrivateRepo"`
	relatedCommands      interface{} `related_commands:"add-plugin-repo, delete-plugin-repo, install-plugin"`
}

func (RepoPluginsCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (RepoPluginsCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
