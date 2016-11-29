package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type UnsharePrivateDomainCommand struct {
	RequiredArgs    flag.OrgDomain `positional-args:"yes"`
	usage           interface{}    `usage:"CF_NAME unshare-private-domain ORG DOMAIN"`
	relatedCommands interface{}    `related_commands:"delete-domain, domains"`
}

func (_ UnsharePrivateDomainCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ UnsharePrivateDomainCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
