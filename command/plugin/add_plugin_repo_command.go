package plugin

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/pluginaction"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/plugin/shared"
)

//go:generate counterfeiter . AddPluginRepoActor

type AddPluginRepoActor interface {
	AddPluginRepository(repoName string, repoURL string) error
}

type AddPluginRepoCommand struct {
	RequiredArgs      flag.AddPluginRepoArgs `positional-args:"yes"`
	usage             interface{}            `usage:"CF_NAME add-plugin-repo REPO_NAME URL\n\nEXAMPLES:\n   CF_NAME add-plugin-repo ExampleRepo https://example.com/repo"`
	relatedCommands   interface{}            `related_commands:"install-plugin, list-plugin-repos"`
	SkipSSLValidation bool                   `short:"k" hidden:"true" description:"Skip SSL certificate validation"`
	UI                command.UI
	Config            command.Config
	Actor             AddPluginRepoActor
}

func (cmd *AddPluginRepoCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.Actor = pluginaction.NewActor(config, shared.NewClient(config, ui, cmd.SkipSSLValidation))
	return nil
}

func (cmd AddPluginRepoCommand) Execute(args []string) error {
	err := cmd.Actor.AddPluginRepository(cmd.RequiredArgs.PluginRepoName, cmd.RequiredArgs.PluginRepoURL)
	switch e := err.(type) {
	case actionerror.RepositoryAlreadyExistsError:
		cmd.UI.DisplayTextWithFlavor("{{.RepositoryURL}} already registered as {{.RepositoryName}}",
			map[string]interface{}{
				"RepositoryName": e.Name,
				"RepositoryURL":  e.URL,
			})
	case nil:
		cmd.UI.DisplayTextWithFlavor("{{.RepositoryURL}} added as {{.RepositoryName}}",
			map[string]interface{}{
				"RepositoryName": cmd.RequiredArgs.PluginRepoName,
				"RepositoryURL":  cmd.RequiredArgs.PluginRepoURL,
			})
	default:
		return err
	}

	return nil
}
