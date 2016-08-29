package v2actions

import (
	"reflect"
	"sort"
	"strings"

	"code.cloudfoundry.org/cli/utils/sortutils"
)

// CommandInfo contains the help details of a command
type CommandInfo struct {
	// Name is the command name
	Name string

	// Description is the command description
	Description string

	// Alias is the command alias
	Alias string

	// Usage is the command usage string, may contain examples and flavor text
	Usage string

	// RelatedCommands is a list of commands related to the command
	RelatedCommands []string

	// Flags contains the list of flags for this command
	Flags []CommandFlag
}

// CommandFlag contains the help details of a command's flag
type CommandFlag struct {
	// Short is the short form of the flag
	Short string

	// Long is the long form of the flag
	Long string

	// Description is the description of the flag
	Description string
}

// GetCommandInfo returns the help information for a particular commandName in the commandList.
func (_ Actor) GetCommandInfo(commandList interface{}, commandName string) (CommandInfo, error) {
	field, found := reflect.TypeOf(commandList).FieldByNameFunc(
		func(fieldName string) bool {
			field, _ := reflect.TypeOf(commandList).FieldByName(fieldName)
			return field.Tag.Get("command") == commandName || field.Tag.Get("alias") == commandName
		},
	)

	if !found {
		return CommandInfo{}, ErrorInvalidCommand{CommandName: commandName}
	}

	tag := field.Tag
	cmd := CommandInfo{
		Name:        tag.Get("command"),
		Description: tag.Get("description"),
		Alias:       tag.Get("alias"),
		Flags:       []CommandFlag{},
	}

	command := field.Type
	for i := 0; i < command.NumField(); i++ {
		fieldTag := command.Field(i).Tag

		if fieldTag.Get("hidden") != "" {
			continue
		}

		if fieldTag.Get("usage") != "" {
			cmd.Usage = fieldTag.Get("usage")
			continue
		}

		if fieldTag.Get("related_commands") != "" {
			relatedCommands := sortutils.Alphabetic(strings.Split(fieldTag.Get("related_commands"), ", "))
			sort.Sort(relatedCommands)
			cmd.RelatedCommands = relatedCommands
			continue
		}

		if fieldTag.Get("description") != "" {
			cmd.Flags = append(cmd.Flags, CommandFlag{
				Short:       fieldTag.Get("short"),
				Long:        fieldTag.Get("long"),
				Description: fieldTag.Get("description"),
			})
		}
	}

	return cmd, nil
}

// GetAllNamesAndDescriptions returns a slice of CommandInfo that only fills in
// the Name and Description for all the commands in commandList
func (_ Actor) GetAllNamesAndDescriptions(commandList interface{}) map[string]CommandInfo {
	handler := reflect.TypeOf(commandList)

	infos := make(map[string]CommandInfo, handler.NumField())
	for i := 0; i < handler.NumField(); i++ {
		fieldTag := handler.Field(i).Tag
		commandName := fieldTag.Get("command")
		infos[commandName] = CommandInfo{
			Name:        commandName,
			Description: fieldTag.Get("description"),
		}
	}

	return infos
}
