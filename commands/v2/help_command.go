package v2

import (
	"fmt"
	"reflect"
	"strings"

	"code.cloudfoundry.org/cli/commands/flags"
	"code.cloudfoundry.org/cli/commands/ui"
)

type HelpCommand struct {
	UI           UI
	OptionalArgs flags.CommandName `positional-args:"yes"`
	usage        interface{}       `usage:"CF_NAME help [COMMAND]"`
}

func (cmd *HelpCommand) Setup() error {
	cmd.UI = ui.NewUI()
	return nil
}

type CommandInfo struct {
	Name        string
	Description string
	Alias       string
	Usage       string
	Flags       []CommandFlags
}

type CommandFlags struct {
	Short       string
	Long        string
	Description string
}

func GetCommandInfo(commandName string) (CommandInfo, error) {
	sanitizedCmdName := strings.ToLower(commandName)
	field, found := reflect.TypeOf(Commands).FieldByNameFunc(
		func(fieldName string) bool {
			field, _ := reflect.TypeOf(Commands).FieldByName(fieldName)
			return field.Tag.Get("command") == sanitizedCmdName || field.Tag.Get("alias") == sanitizedCmdName
		},
	)

	if !found {
		return CommandInfo{}, fmt.Errorf("'%s' is not a registered command. See 'cf help'", commandName)
	}

	tag := field.Tag
	cmd := CommandInfo{
		Name:        tag.Get("command"),
		Description: tag.Get("description"),
		Alias:       tag.Get("alias"),
		Flags:       []CommandFlags{},
	}

	command := field.Type
	for i := 0; i < command.NumField(); i++ {
		fieldTag := command.Field(i).Tag

		if fieldTag.Get("hidden") != "" {
			continue
		}

		if fieldTag.Get("usage") != "" {
			executableName := "cf" //TODO: Figure out how to dynamically get this name
			cmd.Usage = strings.Replace(fieldTag.Get("usage"), "CF_NAME", executableName, -1)
			continue
		}

		if fieldTag.Get("description") != "" {
			cmd.Flags = append(cmd.Flags, CommandFlags{
				Short:       fieldTag.Get("short"),
				Long:        fieldTag.Get("long"),
				Description: fieldTag.Get("description"),
			})
		}
	}

	return cmd, nil
}

func (cmd HelpCommand) Execute(args []string) error {
	cmdInfo, err := GetCommandInfo(cmd.OptionalArgs.CommandName)
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

func (_ HelpCommand) getLongestWidth(flags []CommandFlags) int {
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
