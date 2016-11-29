package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type AddPluginRepoCommand struct {
	RequiredArgs    flag.AddPluginRepoArgs `positional-args:"yes"`
	usage           interface{}            `usage:"CF_NAME add-plugin-repo REPO_NAME URL\n\nEXAMPLES:\n   CF_NAME add-plugin-repo PrivateRepo https://myprivaterepo.com/repo/"`
	relatedCommands interface{}            `related_commands:"install-plugin, list-plugin-repos"`
}

func (_ AddPluginRepoCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ AddPluginRepoCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
