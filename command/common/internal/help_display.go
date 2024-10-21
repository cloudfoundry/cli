package internal

import (
	"fmt"
	"sort"
	"strings"

	"code.cloudfoundry.org/cli/v7/actor/sharedaction"
	"code.cloudfoundry.org/cli/v7/util/configv3"
	"code.cloudfoundry.org/cli/v7/util/sorting"
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

	flagNames := make([]string, 0, len(plugin.UsageDetails.Options))
	for flag := range plugin.UsageDetails.Options {
		flagNames = append(flagNames, flag)
	}
	sort.Slice(flagNames, sorting.SortAlphabeticFunc(flagNames))

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
	for name := range cmds {
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
		name := FlagWithHyphens(flag)

		if len(name) > longest {
			longest = len(name)
		}
	}
	return longest
}

func FlagWithHyphens(flag sharedaction.CommandFlag) string {
	switch {
	case flag.Short != "" && flag.Long != "":
		return fmt.Sprintf("--%s, -%s", flag.Long, flag.Short)
	case flag.Short != "":
		return "-" + flag.Short
	default:
		return "--" + flag.Long
	}
}
