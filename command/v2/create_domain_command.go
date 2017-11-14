package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type CreateDomainCommand struct {
	RequiredArgs    flag.OrgDomain `positional-args:"yes"`
	usage           interface{}    `usage:"CF_NAME create-domain ORG DOMAIN"`
	relatedCommands interface{}    `related_commands:"create-shared-domain, domains, router-groups, share-private-domain"`
}

func (CreateDomainCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (CreateDomainCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
