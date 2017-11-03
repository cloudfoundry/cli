package plugin

import (
	"strings"

	"code.cloudfoundry.org/cli/actor/pluginaction"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/plugin/shared"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
)

//go:generate counterfeiter . PluginsActor

type PluginsActor interface {
	GetOutdatedPlugins() ([]pluginaction.OutdatedPlugin, error)
}

type PluginsCommand struct {
	Checksum          bool        `long:"checksum" description:"Compute and show the sha1 value of the plugin binary file"`
	Outdated          bool        `long:"outdated" description:"Search the plugin repositories for new versions of installed plugins"`
	usage             interface{} `usage:"CF_NAME plugins [--checksum | --outdated]"`
	relatedCommands   interface{} `related_commands:"install-plugin, repo-plugins, uninstall-plugin"`
	SkipSSLValidation bool        `short:"k" hidden:"true" description:"Skip SSL certificate validation"`
	UI                command.UI
	Config            command.Config
	Actor             PluginsActor
}

func (cmd *PluginsCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	pluginClient := shared.NewClient(config, ui, cmd.SkipSSLValidation)
	cmd.Actor = pluginaction.NewActor(config, pluginClient)
	return nil
}

func (cmd PluginsCommand) Execute([]string) error {
	switch {
	case cmd.Outdated:
		return cmd.displayOutdatedPlugins()
	case cmd.Checksum:
		return cmd.displayPluginChecksums(cmd.Config.Plugins())
	default:
		return cmd.displayPluginCommands(cmd.Config.Plugins())
	}
}

func (cmd PluginsCommand) displayPluginChecksums(plugins []configv3.Plugin) error {
	cmd.UI.DisplayText("Computing sha1 for installed plugins, this may take a while...")
	table := [][]string{{"plugin", "version", "sha1"}}
	for _, plugin := range plugins {
		table = append(table, []string{plugin.Name, plugin.Version.String(), plugin.CalculateSHA1()})
	}

	cmd.UI.DisplayNewline()
	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)
	return nil
}

func (cmd PluginsCommand) displayOutdatedPlugins() error {
	repos := cmd.Config.PluginRepositories()
	if len(repos) == 0 {
		return translatableerror.NoPluginRepositoriesError{}
	}
	repoNames := make([]string, len(repos))
	for i := range repos {
		repoNames[i] = repos[i].Name
	}
	cmd.UI.DisplayTextWithFlavor("Searching {{.RepoNames}} for newer versions of installed plugins...",
		map[string]interface{}{
			"RepoNames": strings.Join(repoNames, ", "),
		})

	outdatedPlugins, err := cmd.Actor.GetOutdatedPlugins()
	if err != nil {
		return err
	}

	table := [][]string{{"plugin", "version", "latest version"}}

	for _, plugin := range outdatedPlugins {
		table = append(table, []string{plugin.Name, plugin.CurrentVersion, plugin.LatestVersion})
	}

	cmd.UI.DisplayNewline()
	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)

	cmd.UI.DisplayNewline()
	cmd.UI.DisplayText("Use '{{.BinaryName}} install-plugin' to update a plugin to the latest version.", map[string]interface{}{
		"BinaryName": cmd.Config.BinaryName(),
	})

	return nil
}

func (cmd PluginsCommand) displayPluginCommands(plugins []configv3.Plugin) error {
	cmd.UI.DisplayText("Listing installed plugins...")
	table := [][]string{{"plugin", "version", "command name", "command help"}}
	for _, plugin := range plugins {
		for _, command := range plugin.PluginCommands() {
			table = append(table, []string{plugin.Name, plugin.Version.String(), command.CommandName(), command.HelpText})
		}
	}
	cmd.UI.DisplayNewline()
	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)

	cmd.UI.DisplayNewline()
	cmd.UI.DisplayText("Use '{{.BinaryName}} repo-plugins' to list plugins in registered repos available to install.",
		map[string]interface{}{
			"BinaryName": cmd.Config.BinaryName(),
		})

	return nil
}
