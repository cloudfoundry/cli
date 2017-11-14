package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type UnsharePrivateDomainCommand struct {
	RequiredArgs    flag.OrgDomain `positional-args:"yes"`
	usage           interface{}    `usage:"CF_NAME unshare-private-domain ORG DOMAIN"`
	relatedCommands interface{}    `related_commands:"delete-domain, domains"`
}

func (UnsharePrivateDomainCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (UnsharePrivateDomainCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
