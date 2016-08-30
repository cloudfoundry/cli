package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type DeleteDomainCommand struct {
	RequiredArgs    flags.Domain `positional-args:"yes"`
	Force           bool         `short:"f" description:"Force deletion without confirmation"`
	usage           interface{}  `usage:"CF_NAME delete-domain DOMAIN [-f]"`
	relatedCommands interface{}  `related_commands:"delete-shared-domain, domains, unshare-private-domain"`
}

func (_ DeleteDomainCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ DeleteDomainCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
