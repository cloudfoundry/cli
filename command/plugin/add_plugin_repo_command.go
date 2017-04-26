package plugin

import (
	"os"

	"code.cloudfoundry.org/cli/actor/pluginaction"
	oldCmd "code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/plugin/shared"
)

//go:generate counterfeiter . AddPluginRepoActor

type AddPluginRepoActor interface {
	AddPluginRepository(repoName string, repoURL string) error
}

type AddPluginRepoCommand struct {
	RequiredArgs    flag.AddPluginRepoArgs `positional-args:"yes"`
	usage           interface{}            `usage:"CF_NAME add-plugin-repo REPO_NAME URL\n\nEXAMPLES:\n   CF_NAME add-plugin-repo ExampleRepo https://example.com/repo"`
	relatedCommands interface{}            `related_commands:"install-plugin, list-plugin-repos"`

	UI     command.UI
	Config command.Config
	Actor  AddPluginRepoActor
}

func (cmd *AddPluginRepoCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	pluginClient := shared.NewClient(config, ui)
	cmd.Actor = pluginaction.NewActor(config, pluginClient)
	return nil
}

func (cmd AddPluginRepoCommand) Execute(args []string) error {
	if cmd.Config.Experimental() == false {
		oldCmd.Main(os.Getenv("CF_TRACE"), os.Args)
		return nil
	}
	cmd.UI.DisplayText(command.ExperimentalWarning)
	cmd.UI.DisplayNewline()

	err := cmd.Actor.AddPluginRepository(cmd.RequiredArgs.PluginRepoName, cmd.RequiredArgs.PluginRepoURL)
	if err != nil {
		return shared.HandleError(err)
	}

	cmd.UI.DisplayTextWithFlavor("{{.RepositoryURL}} added as '{{.RepositoryName}}'",
		map[string]interface{}{
			"RepositoryName": cmd.RequiredArgs.PluginRepoName,
			"RepositoryURL":  cmd.RequiredArgs.PluginRepoURL,
		})

	return nil
}
