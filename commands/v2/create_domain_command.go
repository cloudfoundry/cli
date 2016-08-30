package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type CreateDomainCommand struct {
	RequiredArgs    flags.OrgDomain `positional-args:"yes"`
	usage           interface{}     `usage:"CF_NAME create-domain ORG DOMAIN"`
	relatedCommands interface{}     `related_commands:"create-shared-domain, domains, router-groups, share-private-domain"`
}

func (_ CreateDomainCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ CreateDomainCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
