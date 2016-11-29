package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type CreateDomainCommand struct {
	RequiredArgs    flag.OrgDomain `positional-args:"yes"`
	usage           interface{}    `usage:"CF_NAME create-domain ORG DOMAIN"`
	relatedCommands interface{}    `related_commands:"create-shared-domain, domains, router-groups, share-private-domain"`
}

func (_ CreateDomainCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ CreateDomainCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
