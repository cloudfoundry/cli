package common

import (
	"fmt"
	"sort"
	"strings"

	"code.cloudfoundry.org/cli/actors/v2actions"
	"code.cloudfoundry.org/cli/utils/configv3"
	"code.cloudfoundry.org/cli/utils/sortutils"
)

type HelpCategory struct {
	CategoryName string
	CommandList  [][]string
}

func ConvertPluginToCommandInfo(plugin configv3.PluginCommand) v2actions.CommandInfo {
	commandInfo := v2actions.CommandInfo{
		Name:        plugin.Name,
		Description: plugin.HelpText,
		Alias:       plugin.Alias,
		Usage:       plugin.UsageDetails.Usage,
		Flags:       []v2actions.CommandFlag{},
	}

	flagNames := sortutils.Alphabetic{}
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
				v2actions.CommandFlag{
					Short:       strippedFlag,
					Description: description,
				})
		default:
			commandInfo.Flags = append(commandInfo.Flags,
				v2actions.CommandFlag{
					Long:        strippedFlag,
					Description: description,
				})
		}
	}

	return commandInfo
}

func LongestCommandName(cmds map[string]v2actions.CommandInfo, pluginCmds []configv3.PluginCommand) int {
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

func LongestFlagWidth(flags []v2actions.CommandFlag) int {
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
