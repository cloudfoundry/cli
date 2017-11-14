package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type SharePrivateDomainCommand struct {
	RequiredArgs    flag.OrgDomain `positional-args:"yes"`
	usage           interface{}    `usage:"CF_NAME share-private-domain ORG DOMAIN"`
	relatedCommands interface{}    `related_commands:"domains, unshare-private-domain"`
}

func (SharePrivateDomainCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (SharePrivateDomainCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
