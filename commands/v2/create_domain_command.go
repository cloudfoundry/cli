package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type CreateDomainCommand struct {
	RequiredArgs flags.OrgDomain `positional-args:"yes"`
}

func (_ CreateDomainCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
