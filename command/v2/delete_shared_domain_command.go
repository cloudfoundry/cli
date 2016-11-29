package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type DeleteSharedDomainCommand struct {
	RequiredArgs    flag.Domain `positional-args:"yes"`
	Force           bool        `short:"f" description:"Force deletion without confirmation"`
	usage           interface{} `usage:"CF_NAME delete-shared-domain DOMAIN [-f]"`
	relatedCommands interface{} `related_commands:"delete-domain, domains"`
}

func (_ DeleteSharedDomainCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ DeleteSharedDomainCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
