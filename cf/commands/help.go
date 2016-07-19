package commands

import (
	"errors"
	"strings"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/pluginconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/help"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"

	. "code.cloudfoundry.org/cli/cf/i18n"
)

type Help struct {
	ui     terminal.UI
	config pluginconfig.PluginConfiguration
}

func init() {
	commandregistry.Register(&Help{})
}

func (cmd *Help) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "help",
		ShortName:   "h",
		Description: T("Show help"),
		Usage: []string{
			T("CF_NAME help [COMMAND]"),
		},
	}
}

func (cmd *Help) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	reqs := []requirements.Requirement{}
	return reqs, nil
}

func (cmd *Help) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.PluginConfig
	return cmd
}

func (cmd *Help) Execute(c flags.FlagContext) error {
	if len(c.Args()) == 0 {
		help.ShowHelp(cmd.ui.Writer(), help.GetHelpTemplate())
	} else {
		cmdName := c.Args()[0]
		if commandregistry.Commands.CommandExists(cmdName) {
			cmd.ui.Say(commandregistry.Commands.CommandUsage(cmdName))
		} else {
			//check plugin commands
			found := false
			for _, meta := range cmd.config.Plugins() {
				for _, c := range meta.Commands {
					if c.Name == cmdName || c.Alias == cmdName {
						output := T("NAME:") + "\n"
						output += "   " + c.Name + " - " + c.HelpText + "\n"

						if c.Alias != "" {
							output += "\n" + T("ALIAS:") + "\n"
							output += "   " + c.Alias + "\n"
						}

						output += "\n" + T("USAGE:") + "\n"
						output += "   " + c.UsageDetails.Usage + "\n"

						if len(c.UsageDetails.Options) > 0 {
							output += "\n" + T("OPTIONS:") + "\n"

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
				return errors.New("'" + cmdName + "' is not a registered command. See 'cf help'")
			}
		}
	}
	return nil
}
