package plugin

import (
	"code.cloudfoundry.org/cli/v8/actor/pluginaction"
	"code.cloudfoundry.org/cli/v8/command"
	"code.cloudfoundry.org/cli/v8/command/plugin/shared"
	"code.cloudfoundry.org/cli/v8/util/configv3"
	"code.cloudfoundry.org/cli/v8/util/ui"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . RepoPluginsActor

type RepoPluginsActor interface {
	GetPluginRepository(repositoryName string) (configv3.PluginRepository, error)
	GetPluginsFromRepositories(repositories []configv3.PluginRepository) ([]pluginaction.RepositoryPlugins, []error)
}

type RepoPluginsCommand struct {
	RegisteredRepository string      `short:"r" description:"Name of a registered repository"`
	usage                interface{} `usage:"CF_NAME repo-plugins [-r REPO_NAME]\n\nEXAMPLES:\n   CF_NAME repo-plugins -r PrivateRepo"`
	relatedCommands      interface{} `related_commands:"add-plugin-repo, remove-plugin-repo, install-plugin"`
	SkipSSLValidation    bool        `short:"k" hidden:"true" description:"Skip SSL certificate validation"`
	UI                   command.UI
	Config               command.Config
	Actor                RepoPluginsActor
}

func (cmd *RepoPluginsCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.Actor = pluginaction.NewActor(config, shared.NewClient(config, ui, cmd.SkipSSLValidation))
	return nil
}

func (cmd RepoPluginsCommand) Execute(args []string) error {
	repositories := cmd.Config.PluginRepositories()

	if cmd.RegisteredRepository == "" {
		cmd.UI.DisplayText("Getting plugins from all repositories ...")
	} else {
		repository, err := cmd.Actor.GetPluginRepository(cmd.RegisteredRepository)
		if err != nil {
			return err
		}
		repositories = []configv3.PluginRepository{repository}
		cmd.UI.DisplayText("Getting plugins from repository '{{.RepositoryName}}' ...",
			map[string]interface{}{
				"RepositoryName": cmd.RegisteredRepository,
			})
	}
	cmd.UI.DisplayNewline()

	repositoryPlugins, errs := cmd.Actor.GetPluginsFromRepositories(repositories)

	for _, repository := range repositoryPlugins {
		cmd.UI.DisplayTextWithBold("Repository: {{.RepositoryName}}",
			map[string]interface{}{
				"RepositoryName": repository.RepositoryName,
			})

		table := [][]string{{"name", "version", "description"}}
		for _, plugin := range repository.Plugins {
			table = append(table, []string{plugin.Name, plugin.Version, plugin.Description})
		}
		cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)
		cmd.UI.DisplayNewline()
	}

	if len(errs) > 0 {
		cmd.UI.DisplayWarning("Logged errors:")
		for _, err := range errs {
			cmd.UI.DisplayWarning("{{.Error}}", map[string]interface{}{"Error": err.Error()})
		}
		cmd.UI.DisplayNewline()
	}

	return nil
}
