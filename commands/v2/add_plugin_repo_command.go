package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type AddPluginRepoCommand struct {
	RequiredArgs    flags.AddPluginRepoArgs `positional-args:"yes"`
	usage           interface{}             `usage:"CF_NAME add-plugin-repo REPO_NAME URL\n\nEXAMPLES:\n   CF_NAME add-plugin-repo PrivateRepo https://myprivaterepo.com/repo/"`
	relatedCommands interface{}             `related_commands:"install-plugin, list-plugin-repos"`
}

func (_ AddPluginRepoCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ AddPluginRepoCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
