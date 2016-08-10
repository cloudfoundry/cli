package v2

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/actors/v2actions"
	"code.cloudfoundry.org/cli/commands/flags"
	"code.cloudfoundry.org/cli/commands/ui"
)

//go:generate counterfeiter . HelpActor

// HelpActor handles the business logic of the help command
type HelpActor interface {
	// GetCommandInfo returns back a help command information for the given
	// command
	GetCommandInfo(interface{}, string) (v2actions.CommandInfo, error)
}

type HelpCommand struct {
	UI    UI
	Actor HelpActor

	OptionalArgs flags.CommandName `positional-args:"yes"`
	usage        interface{}       `usage:"CF_NAME help [COMMAND]"`
}

func (cmd *HelpCommand) Setup() error {
	cmd.UI = ui.NewUI()
	cmd.Actor = v2actions.NewActor()
	return nil
}

func (cmd HelpCommand) Execute(args []string) error {
	cmdInfo, err := cmd.Actor.GetCommandInfo(Commands, cmd.OptionalArgs.CommandName)
	if err != nil {
		return err
	}

	cmd.UI.DisplayText("NAME:")
	cmd.UI.DisplayTextWithKeyTranslations("    {{.CommandName}} - {{.CommandDescription}}",
		[]string{"CommandDescription"},
		map[string]interface{}{
			"CommandName":        cmdInfo.Name,
			"CommandDescription": cmdInfo.Description,
		})
	cmd.UI.DisplayText("")

	cmd.UI.DisplayText("USAGE:")
	cmd.UI.DisplayTextWithKeyTranslations("    {{.CommandUsage}}",
		[]string{"CommandUsage"},
		map[string]interface{}{
			"CommandUsage": cmdInfo.Usage,
		})
	cmd.UI.DisplayText("")

	if cmdInfo.Alias != "" {
		cmd.UI.DisplayText("ALIAS:")
		cmd.UI.DisplayText("    {{.Alias}}",
			map[string]interface{}{
				"Alias": cmdInfo.Alias,
			})
		cmd.UI.DisplayText("")
	}

	if len(cmdInfo.Flags) != 0 {
		cmd.UI.DisplayText("OPTIONS:")
		nameWidth := cmd.getLongestWidth(cmdInfo.Flags) + 6
		for _, flag := range cmdInfo.Flags {
			var name string
			if flag.Short != "" && flag.Long != "" {
				name = fmt.Sprintf("--%s, -%s", flag.Long, flag.Short)
			} else if flag.Short != "" {
				name = "-" + flag.Short
			} else {
				name = "--" + flag.Long
			}

			cmd.UI.DisplayTextWithKeyTranslations("    {{.Flags}}{{.Spaces}}{{.Description}}",
				[]string{"Description"},
				map[string]interface{}{
					"Flags":       name,
					"Spaces":      strings.Repeat(" ", nameWidth-len(name)),
					"Description": flag.Description,
				})
		}
	}
	return nil
}

func (_ HelpCommand) getLongestWidth(flags []v2actions.CommandFlag) int {
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
