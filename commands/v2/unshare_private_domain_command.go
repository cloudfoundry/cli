package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type UnsharePrivateDomainCommand struct {
	RequiredArgs    flags.OrgDomain `positional-args:"yes"`
	usage           interface{}     `usage:"CF_NAME unshare-private-domain ORG DOMAIN"`
	relatedCommands interface{}     `related_commands:"delete-domain, domains"`
}

func (_ UnsharePrivateDomainCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ UnsharePrivateDomainCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
