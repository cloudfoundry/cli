package commands

import (
	"strings"

	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/plugin_config"
	"github.com/cloudfoundry/cli/cf/help"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"

	. "github.com/cloudfoundry/cli/cf/i18n"
)

type Help struct {
	ui     terminal.UI
	config plugin_config.PluginConfiguration
}

func init() {
	command_registry.Register(&Help{})
}

func (cmd *Help) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "help",
		ShortName:   "h",
		Description: T("Show help"),
		Usage:       T("CF_NAME help [COMMAND]"),
	}
}

func (cmd *Help) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	return
}

func (cmd *Help) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.PluginConfig
	return cmd
}

func (cmd *Help) Execute(c flags.FlagContext) {
	if len(c.Args()) == 0 {
		help.ShowHelp(help.GetHelpTemplate())
	} else {
		cmdName := c.Args()[0]
		if command_registry.Commands.CommandExists(cmdName) {
			cmd.ui.Say(command_registry.Commands.CommandUsage(cmdName))
		} else {
			//check plugin commands
			found := false
			for _, meta := range cmd.config.Plugins() {
				for _, c := range meta.Commands {
					if c.Name == cmdName || c.Alias == cmdName {
						output := T("NAME") + ":" + "\n"
						output += "   " + c.Name + " - " + c.HelpText + "\n"

						if c.Alias != "" {
							output += "\n" + T("ALIAS") + ":" + "\n"
							output += "   " + c.Alias + "\n"
						}

						output += "\n" + T("USAGE") + ":" + "\n"
						output += "   " + c.UsageDetails.Usage + "\n"

						if len(c.UsageDetails.Options) > 0 {
							output += "\n" + T("OPTIONS") + ":" + "\n"

							//find longest name length
							l := 0
							for n := range c.UsageDetails.Options {
								if len(n) > l {
									l = len(n)
								}
							}

							for n, f := range c.UsageDetails.Options {
								output += "   -" + n + strings.Repeat(" ", 7+(l-len(n))) + f + "\n"
							}
						}

						cmd.ui.Say(output)

						found = true
					}
				}
			}

			if !found {
				cmd.ui.Failed("'" + cmdName + "' is not a registered command. See 'cf help'")
			}
		}
	}
}
