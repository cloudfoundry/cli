package v2

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"code.cloudfoundry.org/cli/actors/v2actions"
	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
	"code.cloudfoundry.org/cli/commands/v2/internal"
	"code.cloudfoundry.org/cli/utils/config"
	"code.cloudfoundry.org/cli/utils/sortutils"
)

//go:generate counterfeiter . HelpActor

// HelpActor handles the business logic of the help command
type HelpActor interface {
	// CommandInfoByName returns back a help command information for the given
	// command
	CommandInfoByName(interface{}, string) (v2actions.CommandInfo, error)

	// CommandInfos returns a list of all commands
	CommandInfos(interface{}) map[string]v2actions.CommandInfo
}

type HelpCommand struct {
	UI     commands.UI
	Actor  HelpActor
	Config commands.Config

	OptionalArgs flags.CommandName `positional-args:"yes"`
	AllCommands  bool              `short:"a" description:"All available CLI commands"`
	usage        interface{}       `usage:"CF_NAME help [COMMAND]"`
}

func (cmd *HelpCommand) Setup(config commands.Config, ui commands.UI) error {
	cmd.Actor = v2actions.NewActor()
	cmd.Config = config
	cmd.UI = ui

	return nil
}

func (cmd HelpCommand) Execute(args []string) error {
	var err error
	if cmd.OptionalArgs.CommandName == "" {
		cmd.displayFullHelp()
	} else {
		err = cmd.displayCommand()
	}

	return err
}

func (cmd HelpCommand) displayFullHelp() {
	if cmd.AllCommands {
		cmd.displayHelpPreamble()
		cmd.displayAllCommands()
		cmd.displayHelpFooter()
	} else {
		cmd.displayCommonCommands()
	}
}

func (cmd HelpCommand) displayHelpPreamble() {
	cmd.UI.DisplayHelpHeader("NAME:")
	cmd.UI.DisplayTextWithKeyTranslations("   {{.CommandName}} - {{.CommandDescription}}",
		[]string{"CommandDescription"},
		map[string]interface{}{
			"CommandName":        cmd.Config.BinaryName(),
			"CommandDescription": "A command line tool to interact with Cloud Foundry",
		})
	cmd.UI.DisplayNewline()

	cmd.UI.DisplayHelpHeader("USAGE:")
	cmd.UI.DisplayTextWithKeyTranslations("   {{.CommandName}} {{.CommandUsage}}",
		[]string{"CommandUsage"},
		map[string]interface{}{
			"CommandName":  cmd.Config.BinaryName(),
			"CommandUsage": "[global options] command [arguments...] [command options]",
		})
	cmd.UI.DisplayNewline()

	cmd.UI.DisplayHelpHeader("VERSION:")
	cmd.UI.DisplayText("   {{.Version}}-{{.Time}}", map[string]interface{}{
		"Version": cf.Version,
		"Time":    cf.BuiltOnDate,
	})
	cmd.UI.DisplayNewline()
}

func (cmd HelpCommand) displayCommonCommands() {
	cmdInfo := cmd.Actor.CommandInfos(Commands)
	prefix := "  "

	cmd.UI.DisplayTextWithKeyTranslations("{{.CommandName}} {{.VersionCommand}} {{.Version}}-{{.Time}}, {{.CLI}}",
		[]string{"VersionCommand", "CLI"},
		map[string]interface{}{
			"CommandName":    cmd.Config.BinaryName(),
			"VersionCommand": "version",
			"Version":        cf.Version,
			"Time":           cf.BuiltOnDate,
			"CLI":            "Cloud Foundry command line tool",
		})
	cmd.UI.DisplayTextWithKeyTranslations("{{.Usage}} {{.CommandName}} {{.CommandUsage}}",
		[]string{"Usage", "CommandUsage"},
		map[string]interface{}{
			"Usage":        "Usage:",
			"CommandName":  cmd.Config.BinaryName(),
			"CommandUsage": "[global options] command [arguments...] [command options]",
		})
	cmd.UI.DisplayNewline()

	for _, category := range internal.CommonHelpCategoryList {
		cmd.UI.DisplayHelpHeader(category.CategoryName)
		table := [][]string{}

		for _, row := range category.CommandList {
			finalRow := []string{}

			for _, command := range row {
				separator := ""
				if info, ok := cmdInfo[command]; ok {
					if len(info.Alias) > 0 {
						separator = ","
					}
					finalRow = append(finalRow, fmt.Sprintf("%s%s%s", info.Name, separator, info.Alias))
				}
			}

			table = append(table, finalRow)
		}

		cmd.UI.DisplayTable(prefix, table)
		cmd.UI.DisplayNewline()
	}

	pluginCommands := cmd.getSortedPluginCommands()
	cmd.UI.DisplayHelpHeader("Commands offered by installed plugins:")

	size := int(math.Ceil(float64(len(pluginCommands)) / 3))
	table := make([][]string, size)
	for i := 0; i < size; i++ {
		table[i] = make([]string, 3)
		for j := 0; j < 3; j++ {
			index := i + j*size
			if index < len(pluginCommands) {
				table[i][j] = pluginCommands[index].Name
			}
		}
	}

	cmd.UI.DisplayTable(prefix, table)
	cmd.UI.DisplayNewline()

	cmd.UI.DisplayHelpHeader("Global options:")
	cmd.UI.DisplayTextWithKeyTranslations(prefix+"{{.ENVName}}                         {{.Description}}",
		[]string{"Description"},
		map[string]interface{}{
			"ENVName":     "--help, -h",
			"Description": "Show help",
		})
	cmd.UI.DisplayTextWithKeyTranslations(prefix+"{{.ENVName}}                                 {{.Description}}",
		[]string{"Description"},
		map[string]interface{}{
			"ENVName":     "-v",
			"Description": "Print API request diagnostics to stdout",
		})
	cmd.UI.DisplayNewline()
	cmd.UI.DisplayText("'cf help -a' lists all commands with short descriptions. See 'cf help <command>' to read about a specific command.")
}

func (cmd HelpCommand) displayAllCommands() {
	pluginCommands := cmd.getSortedPluginCommands()
	cmdInfo := cmd.Actor.CommandInfos(Commands)
	longestCmd := internal.LongestCommandName(cmdInfo, pluginCommands)

	for _, category := range internal.HelpCategoryList {
		cmd.UI.DisplayHelpHeader(category.CategoryName)

		for _, row := range category.CommandList {
			for _, command := range row {
				cmd.UI.DisplayTextWithKeyTranslations("   {{.CommandName}}{{.Gap}}{{.CommandDescription}}",
					[]string{"CommandDescription"},
					map[string]interface{}{
						"CommandName":        cmdInfo[command].Name,
						"CommandDescription": cmdInfo[command].Description,
						"Gap":                strings.Repeat(" ", longestCmd+1-len(command)),
					})
			}

			cmd.UI.DisplayNewline()
		}
	}

	cmd.UI.DisplayHelpHeader("INSTALLED PLUGIN COMMANDS:")
	for _, pluginCommand := range pluginCommands {
		cmd.UI.DisplayText("   {{.CommandName}}{{.Gap}}{{.CommandDescription}}", map[string]interface{}{
			"CommandName":        pluginCommand.Name,
			"CommandDescription": pluginCommand.HelpText,
			"Gap":                strings.Repeat(" ", longestCmd+1-len(pluginCommand.Name)),
		})
	}
	cmd.UI.DisplayNewline()
}

func (cmd HelpCommand) displayHelpFooter() {
	cmd.UI.DisplayHelpHeader("ENVIRONMENT VARIABLES:")
	cmd.UI.DisplayTextWithKeyTranslations("   {{.ENVName}}                     {{.Description}}",
		[]string{"Description"},
		map[string]interface{}{
			"ENVName":     "CF_COLOR=false",
			"Description": "Do not colorize output",
		})
	cmd.UI.DisplayTextWithKeyTranslations("   {{.ENVName}}               {{.Description}}",
		[]string{"Description"},
		map[string]interface{}{
			"ENVName":     "CF_HOME=path/to/dir/",
			"Description": "Override path to default config directory",
		})
	cmd.UI.DisplayTextWithKeyTranslations("   {{.ENVName}}        {{.Description}}",
		[]string{"Description"},
		map[string]interface{}{
			"ENVName":     "CF_PLUGIN_HOME=path/to/dir/",
			"Description": "Override path to default plugin config directory",
		})
	cmd.UI.DisplayTextWithKeyTranslations("   {{.ENVName}}                      {{.Description}}",
		[]string{"Description"},
		map[string]interface{}{
			"ENVName":     "CF_TRACE=true",
			"Description": "Print API request diagnostics to stdout",
		})
	cmd.UI.DisplayTextWithKeyTranslations("   {{.ENVName}}         {{.Description}}",
		[]string{"Description"},
		map[string]interface{}{
			"ENVName":     "CF_TRACE=path/to/trace.log",
			"Description": "Append API request diagnostics to a log file",
		})
	cmd.UI.DisplayTextWithKeyTranslations("   {{.ENVName}} {{.Description}}",
		[]string{"Description"},
		map[string]interface{}{
			"ENVName":     "https_proxy=proxy.example.com:8080",
			"Description": "Enable HTTP proxying for API requests",
		})
	cmd.UI.DisplayNewline()

	cmd.UI.DisplayHelpHeader("GLOBAL OPTIONS:")
	cmd.UI.DisplayTextWithKeyTranslations("   {{.ENVName}}                         {{.Description}}",
		[]string{"Description"},
		map[string]interface{}{
			"ENVName":     "--help, -h",
			"Description": "Show help",
		})
	cmd.UI.DisplayTextWithKeyTranslations("   {{.ENVName}}                                 {{.Description}}",
		[]string{"Description"},
		map[string]interface{}{
			"ENVName":     "-v",
			"Description": "Print API request diagnostics to stdout",
		})
}

func (cmd HelpCommand) displayCommand() error {
	cmdInfo, err := cmd.Actor.CommandInfoByName(Commands, cmd.OptionalArgs.CommandName)
	if err != nil {
		if err, ok := err.(v2actions.ErrorInvalidCommand); ok {
			var found bool
			if cmdInfo, found = cmd.findPlugin(); !found {
				return err
			}
		} else {
			return err
		}
	}

	cmd.UI.DisplayText("NAME:")
	cmd.UI.DisplayTextWithKeyTranslations("   {{.CommandName}} - {{.CommandDescription}}",
		[]string{"CommandDescription"},
		map[string]interface{}{
			"CommandName":        cmdInfo.Name,
			"CommandDescription": cmdInfo.Description,
		})

	cmd.UI.DisplayNewline()
	usageString := strings.Replace(cmdInfo.Usage, "CF_NAME", cmd.Config.BinaryName(), -1)
	cmd.UI.DisplayText("USAGE:")
	cmd.UI.DisplayTextWithKeyTranslations("   {{.CommandUsage}}",
		[]string{"CommandUsage"},
		map[string]interface{}{
			"CommandUsage": usageString,
		})

	if cmdInfo.Alias != "" {
		cmd.UI.DisplayNewline()
		cmd.UI.DisplayText("ALIAS:")
		cmd.UI.DisplayText("   {{.Alias}}",
			map[string]interface{}{
				"Alias": cmdInfo.Alias,
			})
	}

	if len(cmdInfo.Flags) != 0 {
		cmd.UI.DisplayNewline()
		cmd.UI.DisplayText("OPTIONS:")
		nameWidth := internal.LongestFlagWidth(cmdInfo.Flags) + 6
		for _, flag := range cmdInfo.Flags {
			var name string
			if flag.Short != "" && flag.Long != "" {
				name = fmt.Sprintf("--%s, -%s", flag.Long, flag.Short)
			} else if flag.Short != "" {
				name = "-" + flag.Short
			} else {
				name = "--" + flag.Long
			}

			cmd.UI.DisplayTextWithKeyTranslations("   {{.Flags}}{{.Spaces}}{{.Description}}",
				[]string{"Description"},
				map[string]interface{}{
					"Flags":       name,
					"Spaces":      strings.Repeat(" ", nameWidth-len(name)),
					"Description": flag.Description,
				})
		}
	}

	if len(cmdInfo.Environment) != 0 {
		cmd.UI.DisplayNewline()
		cmd.UI.DisplayText("ENVIRONMENT:")
		for _, envVar := range cmdInfo.Environment {
			cmd.UI.DisplayTextWithKeyTranslations("   {{.EnvVar}}{{.Description}}",
				[]string{"Description"},
				map[string]interface{}{
					"EnvVar":      fmt.Sprintf("%-29s", fmt.Sprintf("%s=%s", envVar.Name, envVar.DefaultValue)),
					"Description": envVar.Description,
				})
		}
	}

	if len(cmdInfo.RelatedCommands) > 0 {
		cmd.UI.DisplayNewline()
		cmd.UI.DisplayText("SEE ALSO:")
		cmd.UI.DisplayText("   {{.RelatedCommands}}",
			map[string]interface{}{
				"RelatedCommands": strings.Join(cmdInfo.RelatedCommands, ", "),
			})
	}

	return nil
}

func (cmd HelpCommand) findPlugin() (v2actions.CommandInfo, bool) {
	for _, pluginConfig := range cmd.Config.Plugins() {
		for _, command := range pluginConfig.Commands {
			if command.Name == cmd.OptionalArgs.CommandName {
				return internal.ConvertPluginToCommandInfo(command), true
			}
		}
	}

	return v2actions.CommandInfo{}, false
}

func (cmd HelpCommand) getSortedPluginCommands() config.PluginCommands {
	plugins := cmd.Config.Plugins()

	sortedPluginNames := sortutils.Alphabetic{}
	for plugin, _ := range plugins {
		sortedPluginNames = append(sortedPluginNames, plugin)
	}
	sort.Sort(sortedPluginNames)

	pluginCommands := config.PluginCommands{}
	for _, pluginName := range sortedPluginNames {
		sortedCommands := plugins[pluginName].Commands
		sort.Sort(sortedCommands)
		pluginCommands = append(pluginCommands, sortedCommands...)
	}

	return pluginCommands
}
