package internal

import (
	"fmt"
	"sort"
	"strings"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/sorting"
)

type HelpCategory struct {
	CategoryName string
	CommandList  [][]string
}

func ConvertPluginToCommandInfo(plugin configv3.PluginCommand) sharedaction.CommandInfo {
	commandInfo := sharedaction.CommandInfo{
		Name:        plugin.Name,
		Description: plugin.HelpText,
		Alias:       plugin.Alias,
		Usage:       plugin.UsageDetails.Usage,
		Flags:       []sharedaction.CommandFlag{},
	}

	flagNames := sorting.Alphabetic{}
	for flag := range plugin.UsageDetails.Options {
		flagNames = append(flagNames, flag)
	}
	sort.Sort(flagNames)

	for _, flag := range flagNames {
		description := plugin.UsageDetails.Options[flag]
		strippedFlag := strings.Trim(flag, "-")
		switch len(flag) {
		case 1:
			commandInfo.Flags = append(commandInfo.Flags,
				sharedaction.CommandFlag{
					Short:       strippedFlag,
					Description: description,
				})
		default:
			commandInfo.Flags = append(commandInfo.Flags,
				sharedaction.CommandFlag{
					Long:        strippedFlag,
					Description: description,
				})
		}
	}

	return commandInfo
}

func LongestCommandName(cmds map[string]sharedaction.CommandInfo, pluginCmds []configv3.PluginCommand) int {
	longest := 0
	for name, _ := range cmds {
		if len(name) > longest {
			longest = len(name)
		}
	}
	for _, command := range pluginCmds {
		if len(command.Name) > longest {
			longest = len(command.Name)
		}
	}
	return longest
}

func LongestFlagWidth(flags []sharedaction.CommandFlag) int {
	longest := 0
	for _, flag := range flags {
		var name string
		if flag.Short != "" && flag.Long != "" {
			name = fmt.Sprintf("--%s, -%s", flag.Long, flag.Short)
		} else if flag.Short != "" {
			name = "-" + flag.Short
		} else {
			name = "--" + flag.Long
		}
		if len(name) > longest {
			longest = len(name)
		}
	}
	return longest
}
