package plugin

import (
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/plugin_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type Plugins struct {
	ui     terminal.UI
	config plugin_config.PluginConfiguration
}

func NewPlugins(ui terminal.UI, config plugin_config.PluginConfiguration) *Plugins {
	return &Plugins{
		ui:     ui,
		config: config,
	}
}

func (cmd *Plugins) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "plugins",
		Description: T("list all available plugin commands"),
		Usage:       T("CF_NAME plugins"),
	}
}

func (cmd *Plugins) GetRequirements(_ requirements.Factory, c *cli.Context) (req []requirements.Requirement, err error) {
	if len(c.Args()) != 0 {
		cmd.ui.FailWithUsage(c)
	}
	return
}

func (cmd *Plugins) Run(c *cli.Context) {
	cmd.ui.Say(T("Listing Installed Plugins..."))

	plugins := cmd.config.Plugins()

	table := terminal.NewTable(cmd.ui, []string{T("Plugin Name"), T("Command Name"), T("Command Help")})

	for pluginName, metadata := range plugins {
		for _, command := range metadata.Commands {
			if command.Alias == "" {
				table.Add(pluginName, command.Name, command.HelpText)
			} else {
				table.Add(pluginName, command.Name+", "+command.Alias, command.HelpText)
			}
		}
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	table.Print()
}
