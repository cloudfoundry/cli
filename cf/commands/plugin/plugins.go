package plugin

import (
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/plugin/rpc"
	"github.com/codegangsta/cli"
)

type Plugins struct {
	ui     terminal.UI
	config configuration.ReadWriter
}

func NewPlugins(ui terminal.UI, config configuration.ReadWriter) *Plugins {
	return &Plugins{
		ui:     ui,
		config: config,
	}
}

func (cmd *Plugins) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "plugins",
		Description: "list all available plugin commands",
		Usage:       "CF_NAME plugins",
	}
}

func (cmd *Plugins) GetRequirements(_ requirements.Factory, _ *cli.Context) (req []requirements.Requirement, err error) {
	return
}

func (cmd *Plugins) Run(c *cli.Context) {
	cmd.ui.Say("Listing Installed Plugins...")
	plugins := cmd.config.Plugins()

	table := terminal.NewTable(cmd.ui, []string{"Plugin name", "Command name"})

	for pluginName, location := range plugins {
		cmdList, _ := rpc.RunListCmd(location)

		for _, command := range cmdList {
			table.Add(pluginName, command.Name)
		}
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	table.Print()
}
