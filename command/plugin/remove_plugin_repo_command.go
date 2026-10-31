package plugin

import (
	"strings"

	"code.cloudfoundry.org/cli/v8/actor/pluginaction"
	"code.cloudfoundry.org/cli/v8/command"
	"code.cloudfoundry.org/cli/v8/command/flag"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . RemovePluginRepoActor

type RemovePluginRepoActor interface {
	RemovePluginRepository(repositoryName string) error
}

type RemovePluginRepoCommand struct {
	RequiredArgs    flag.PluginRepoName `positional-args:"yes"`
	usage           interface{}         `usage:"CF_NAME remove-plugin-repo REPO_NAME\n\nEXAMPLES:\n   CF_NAME remove-plugin-repo PrivateRepo"`
	relatedCommands interface{}         `related_commands:"list-plugin-repos"`
	UI              command.UI
	Config          command.Config
	Actor           RemovePluginRepoActor
}

func (cmd *RemovePluginRepoCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.Actor = pluginaction.NewActor(config, nil)
	return nil
}

func (cmd RemovePluginRepoCommand) Execute(args []string) error {
	repositoryName := strings.TrimSpace(cmd.RequiredArgs.PluginRepoName)

	err := cmd.Actor.RemovePluginRepository(repositoryName)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayText("{{.RepositoryName}} removed from list of repositories",
		map[string]interface{}{
			"RepositoryName": repositoryName,
		})
	return nil
}
