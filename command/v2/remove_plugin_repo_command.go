package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flags"
)

type RemovePluginRepoCommand struct {
	RequiredArgs    flags.PluginRepoName `positional-args:"yes"`
	usage           interface{}          `usage:"CF_NAME remove-plugin-repo REPO_NAME\n\nEXAMPLES:\n   CF_NAME remove-plugin-repo PrivateRepo"`
	relatedCommands interface{}          `related_commands:"list-plugin-repos"`
}

func (_ RemovePluginRepoCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ RemovePluginRepoCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
