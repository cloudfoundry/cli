package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type MarketplaceCommand struct {
	ServicePlanInfo string      `short:"s" description:"Show plan details for a particular service offering"`
	usage           interface{} `usage:"CF_NAME marketplace [-s SERVICE]"`
	relatedCommands interface{} `related_commands:"create-service, services"`
}

func (MarketplaceCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (MarketplaceCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
