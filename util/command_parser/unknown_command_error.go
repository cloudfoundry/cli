package command_parser

import (
	"fmt"
	"reflect"

	"code.cloudfoundry.org/cli/cf/util/spellcheck"
	"code.cloudfoundry.org/cli/command/common"
)

type UnknownCommandError struct {
	CommandName string
	suggestions []string
}

func (e *UnknownCommandError) Suggest(pluginCommandNames []string) {
	var commandNames []string

	commandListStruct := reflect.TypeOf(common.Commands)
	commandListStruct.FieldByNameFunc(
		func(fieldName string) bool {
			field, _ := commandListStruct.FieldByName(fieldName)
			if commandName := field.Tag.Get("command"); commandName != "" {
				commandNames = append(commandNames, commandName)
			}
			return false
		})

	cmdSuggester := spellcheck.NewCommandSuggester(append(commandNames, pluginCommandNames...))
	e.suggestions = cmdSuggester.Recommend(e.CommandName)
}

func (e UnknownCommandError) Error() string {
	message := fmt.Sprintf("'%s' is not a registered command. See 'cf help -a'", e.CommandName)

	if len(e.suggestions) > 0 {
		message += "\n\nDid you mean?"

		for _, suggestion := range e.suggestions {
			message += "\n    " + suggestion
		}
	}

	return message
}
