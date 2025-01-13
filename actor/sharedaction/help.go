package sharedaction

import (
	"reflect"
	"sort"
	"strings"

	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/util/sorting"
)

const (
	CommonCommandsIndent string = "  "
	AllCommandsIndent    string = "   "
	CommandIndent        string = "   "
)

// CommandInfo contains the help details of a command
type CommandInfo struct {
	// Name is the command name
	Name string

	// Description is the command description
	Description string

	// Alias is the command alias
	Alias string

	// Usage is the command usage string
	Usage string

	// Examples is the command examples string
	Examples string

	// Resources is the types of object that the command applies to
	Resources string

	// RelatedCommands is a list of commands related to the command
	RelatedCommands []string

	// Flags contains the list of flags for this command
	Flags []CommandFlag

	// Environment is a list of environment variables specific for this command
	Environment []EnvironmentVariable
}

// CommandFlag contains the help details of a command's flag
type CommandFlag struct {
	// Short is the short form of the flag
	Short string

	// Long is the long form of the flag
	Long string

	// Description is the description of the flag
	Description string

	// Default is the flag's default value
	Default string
}

// EnvironmentVariable contains env vars specific for this command
type EnvironmentVariable struct {
	Name         string
	Description  string
	DefaultValue string
}

// HasUsage is an interface that commands may implement if they want to define their usage
// text in a Usage() method, which gives them more flexibility than a struct tag.
type HasUsage interface {
	Usage() string
}

// HasExamples is an interface that commands may implement if they want to define their examples
// text in a Examples() method, which gives them more flexibility than a struct tag.
type HasExamples interface {
	Examples() string
}

// HasResources is an interface that commands may implement if they want to define their resources
// text in a Resources() method, which gives them more flexibility than a struct tag.
type HasResources interface {
	Resources() string
}

// CommandInfoByName returns the help information for a particular commandName in
// the commandList.
func (Actor) CommandInfoByName(commandList interface{}, commandName string) (CommandInfo, error) {
	field, found := reflect.TypeOf(commandList).FieldByNameFunc(
		func(fieldName string) bool {
			field, _ := reflect.TypeOf(commandList).FieldByName(fieldName)
			return field.Tag.Get("command") == commandName || field.Tag.Get("alias") == commandName
		},
	)

	if !found {
		return CommandInfo{}, actionerror.InvalidCommandError{CommandName: commandName}
	}

	tag := field.Tag
	cmd := CommandInfo{
		Name:        tag.Get("command"),
		Description: tag.Get("description"),
		Alias:       tag.Get("alias"),
		Flags:       []CommandFlag{},
		Environment: []EnvironmentVariable{},
	}

	fieldValue := reflect.ValueOf(commandList).FieldByIndex(field.Index)

	if commandWithUsage, hasUsage := fieldValue.Interface().(HasUsage); hasUsage {
		cmd.Usage = strings.ReplaceAll(
			strings.TrimSpace(commandWithUsage.Usage()),
			"\n",
			"\n"+CommandIndent,
		)
	}

	if commandWithExamples, hasExamples := fieldValue.Interface().(HasExamples); hasExamples {
		cmd.Examples = strings.ReplaceAll(
			strings.TrimSpace(commandWithExamples.Examples()),
			"\n",
			"\n"+CommandIndent,
		)
	}

	if commandWithResources, hasResources := fieldValue.Interface().(HasResources); hasResources {
		cmd.Resources = strings.ReplaceAll(
			strings.TrimSpace(commandWithResources.Resources()),
			"\n",
			"\n"+CommandIndent,
		)
	}

	command := field.Type
	for i := 0; i < command.NumField(); i++ {
		fieldTag := command.Field(i).Tag

		if fieldTag.Get("hidden") != "" {
			continue
		}

		if cmd.Usage == "" && fieldTag.Get("usage") != "" {
			cmd.Usage = fieldTag.Get("usage")
			continue
		}

		if cmd.Examples == "" && fieldTag.Get("examples") != "" {
			cmd.Examples = fieldTag.Get("examples")
			continue
		}

		if fieldTag.Get("related_commands") != "" {
			relatedCommands := strings.Split(fieldTag.Get("related_commands"), ", ")
			sort.Slice(relatedCommands, sorting.SortAlphabeticFunc(relatedCommands))
			cmd.RelatedCommands = relatedCommands
			continue
		}

		if fieldTag.Get("description") != "" {
			cmd.Flags = append(cmd.Flags, CommandFlag{
				Short:       fieldTag.Get("short"),
				Long:        fieldTag.Get("long"),
				Description: fieldTag.Get("description"),
				Default:     fieldTag.Get("default"),
			})
		}

		if fieldTag.Get("environmentName") != "" {
			cmd.Environment = append(cmd.Environment, EnvironmentVariable{
				Name:         fieldTag.Get("environmentName"),
				DefaultValue: fieldTag.Get("environmentDefault"),
				Description:  fieldTag.Get("environmentDescription"),
			})
		}
	}

	return cmd, nil
}

// CommandInfos returns a slice of CommandInfo that only fills in
// the Name and Description for all the commands in commandList
func (Actor) CommandInfos(commandList interface{}) map[string]CommandInfo {
	handler := reflect.TypeOf(commandList)

	infos := make(map[string]CommandInfo, handler.NumField())
	for i := 0; i < handler.NumField(); i++ {
		fieldTag := handler.Field(i).Tag
		commandName := fieldTag.Get("command")
		infos[commandName] = CommandInfo{
			Name:        commandName,
			Description: fieldTag.Get("description"),
			Alias:       fieldTag.Get("alias"),
		}
	}

	return infos
}
