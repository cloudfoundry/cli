package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type DeleteCommand struct {
	RequiredArgs       flag.AppName `positional-args:"yes"`
	ForceDelete        bool         `short:"f" description:"Force deletion without confirmation"`
	DeleteMappedRoutes bool         `short:"r" description:"Also delete any mapped routes"`
	usage              interface{}  `usage:"CF_NAME delete APP_NAME [-r] [-f]"`
	relatedCommands    interface{}  `related_commands:"apps, scale, stop"`
}

func (DeleteCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (DeleteCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
