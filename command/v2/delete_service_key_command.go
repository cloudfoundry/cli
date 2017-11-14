package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type DeleteServiceKeyCommand struct {
	RequiredArgs    flag.ServiceInstanceKey `positional-args:"yes"`
	Force           bool                    `short:"f" description:"Force deletion without confirmation"`
	usage           interface{}             `usage:"CF_NAME delete-service-key SERVICE_INSTANCE SERVICE_KEY [-f]\n\nEXAMPLES:\n   CF_NAME delete-service-key mydb mykey"`
	relatedCommands interface{}             `related_commands:"service-keys"`
}

func (DeleteServiceKeyCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (DeleteServiceKeyCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
