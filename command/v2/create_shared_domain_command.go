package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type CreateSharedDomainCommand struct {
	RequiredArgs    flag.Domain `positional-args:"yes"`
	RouterGroup     string      `long:"router-group" description:"Routes for this domain will be configured only on the specified router group"`
	usage           interface{} `usage:"CF_NAME create-shared-domain DOMAIN [--router-group ROUTER_GROUP]"`
	relatedCommands interface{} `related_commands:"create-domain, domains, router-groups"`
}

func (CreateSharedDomainCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (CreateSharedDomainCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
