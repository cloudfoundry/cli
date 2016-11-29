package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type CreateSharedDomainCommand struct {
	RequiredArgs    flag.Domain `positional-args:"yes"`
	RouterGroup     string      `long:"router-group" description:"Routes for this domain will be configured only on the specified router group"`
	usage           interface{} `usage:"CF_NAME create-shared-domain DOMAIN [--router-group ROUTER_GROUP]"`
	relatedCommands interface{} `related_commands:"create-domain, domains, router-groups"`
}

func (_ CreateSharedDomainCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ CreateSharedDomainCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
