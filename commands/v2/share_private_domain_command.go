package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type SharePrivateDomainCommand struct {
	RequiredArgs flags.OrgDomain `positional-args:"yes"`
}

func (_ SharePrivateDomainCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
