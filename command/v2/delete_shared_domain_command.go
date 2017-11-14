package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type DeleteSharedDomainCommand struct {
	RequiredArgs    flag.Domain `positional-args:"yes"`
	Force           bool        `short:"f" description:"Force deletion without confirmation"`
	usage           interface{} `usage:"CF_NAME delete-shared-domain DOMAIN [-f]"`
	relatedCommands interface{} `related_commands:"delete-domain, domains"`
}

func (DeleteSharedDomainCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (DeleteSharedDomainCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
