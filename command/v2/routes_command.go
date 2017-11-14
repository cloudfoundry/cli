package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type RoutesCommand struct {
	OrgLevel        bool        `long:"orglevel" description:"List all the routes for all spaces of current organization"`
	usage           interface{} `usage:"CF_NAME routes [--orglevel]"`
	relatedCommands interface{} `related_commands:"check-route, domains, map-route, unmap-route"`
}

func (RoutesCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (RoutesCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
