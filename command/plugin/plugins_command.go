package plugin

import (
	"os"
	"sort"

	oldCmd "code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/sorting"
)

type PluginsCommand struct {
	Checksum        bool        `long:"checksum" description:"Compute and show the sha1 value of the plugin binary file"`
	usage           interface{} `usage:"CF_NAME plugins [--checksum]"`
	relatedCommands interface{} `related_commands:"install-plugin, repo-plugins, uninstall-plugin"`

	UI     command.UI
	Config command.Config
}

func (cmd *PluginsCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	return nil
}

func (cmd PluginsCommand) Execute(_ []string) error {
	if cmd.Config.Experimental() == false {
		oldCmd.Main(os.Getenv("CF_TRACE"), os.Args)
		return nil
	}
	cmd.UI.DisplayText(command.ExperimentalWarning)
	cmd.UI.DisplayNewline()

	plugins := cmd.Config.Plugins()
	sortedPluginNames := sorting.Alphabetic{}
	for pluginName, _ := range plugins {
		sortedPluginNames = append(sortedPluginNames, pluginName)
	}
	sort.Sort(sortedPluginNames)

	var table [][]string
	if cmd.Checksum {
		cmd.UI.DisplayText("Computing sha1 for installed plugins, this may take a while...")
		table = cmd.calculatePluginChecksums(sortedPluginNames, plugins)
	} else {
		cmd.UI.DisplayText("Listing installed plugins...")
		table = cmd.getPluginCommands(sortedPluginNames, plugins)
	}

	cmd.UI.DisplayNewline()
	cmd.UI.DisplayTableWithHeader("", table, 3)
	return nil
}

func (cmd PluginsCommand) calculatePluginChecksums(sortedPluginNames []string, plugins map[string]configv3.Plugin) [][]string {
	table := [][]string{{"plugin name", "version", "sha1"}}
	for _, pluginName := range sortedPluginNames {
		plugin := plugins[pluginName]
		table = append(table, []string{pluginName, plugin.Version.String(), plugin.CalculateSHA1()})
	}
	return table
}

func (cmd PluginsCommand) getPluginCommands(sortedPluginNames []string, plugins map[string]configv3.Plugin) [][]string {
	table := [][]string{{"plugin name", "version", "command name", "command help"}}
	for _, pluginName := range sortedPluginNames {
		plugin := plugins[pluginName]
		for _, command := range plugin.Commands {
			table = append(table, []string{pluginName, plugin.Version.String(), command.CommandName(), command.HelpText})
		}
	}
	return table
}
