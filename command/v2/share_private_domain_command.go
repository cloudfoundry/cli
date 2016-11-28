package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flags"
)

type SharePrivateDomainCommand struct {
	RequiredArgs    flags.OrgDomain `positional-args:"yes"`
	usage           interface{}     `usage:"CF_NAME share-private-domain ORG DOMAIN"`
	relatedCommands interface{}     `related_commands:"domains, unshare-private-domain"`
}

func (_ SharePrivateDomainCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ SharePrivateDomainCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
