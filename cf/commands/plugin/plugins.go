package plugin

import (
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/plugin_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/plugin/rpc"
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

func (cmd *Plugins) GetRequirements(_ requirements.Factory, _ *cli.Context) (req []requirements.Requirement, err error) {
	return
}

func (cmd *Plugins) Run(c *cli.Context) {
	cmd.ui.Say(T("Listing Installed Plugins..."))

	plugins := cmd.config.Plugins()

	table := terminal.NewTable(cmd.ui, []string{T("Plugin name"), T("Command name")})

	service, err := rpc.NewRpcService()
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	err = service.Start()
	if err != nil {
		cmd.ui.Failed(err.Error())
	}
	defer service.Stop()

	for pluginName, location := range plugins {
		cmdList, _ := rpc.RunListCmd(pluginName, location, service.Port())

		for _, command := range cmdList {
			table.Add(pluginName, command.Name)
		}
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	table.Print()
}
