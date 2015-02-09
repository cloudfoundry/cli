package plugin

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/plugin_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/utils"
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
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "checksum", Usage: T("Compute and show the sha1 value of the plugin binary file")},
		},
	}
}

func (cmd *Plugins) GetRequirements(_ requirements.Factory, c *cli.Context) (req []requirements.Requirement, err error) {
	if len(c.Args()) != 0 {
		cmd.ui.FailWithUsage(c)
	}
	return
}

func (cmd *Plugins) Run(c *cli.Context) {
	var version string

	cmd.ui.Say(T("Listing Installed Plugins..."))

	plugins := cmd.config.Plugins()

	var table terminal.Table
	if c.Bool("checksum") {
		cmd.ui.Say(T("Computing sha1 for installed plugins, this may take a while ..."))
		table = terminal.NewTable(cmd.ui, []string{T("Plugin Name"), T("Version"), T("Command Name"), "sha1", T("Command Help")})
	} else {
		table = terminal.NewTable(cmd.ui, []string{T("Plugin Name"), T("Version"), T("Command Name"), T("Command Help")})
	}

	for pluginName, metadata := range plugins {
		if metadata.Version.Major == 0 && metadata.Version.Minor == 0 && metadata.Version.Build == 0 {
			version = "N/A"
		} else {
			version = fmt.Sprintf("%d.%d.%d", metadata.Version.Major, metadata.Version.Minor, metadata.Version.Build)
		}

		for _, command := range metadata.Commands {
			args := []string{pluginName, version}

			if command.Alias != "" {
				args = append(args, command.Name+", "+command.Alias)
			} else {
				args = append(args, command.Name)
			}

			if c.Bool("checksum") {
				checksum := utils.NewSha1Checksum(metadata.Location)
				sha1, err := checksum.ComputeFileSha1()
				if err != nil {
					args = append(args, "n/a")
				} else {
					args = append(args, fmt.Sprintf("%x", sha1))
				}
			}

			args = append(args, command.HelpText)
			table.Add(args...)
		}
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	table.Print()
}
