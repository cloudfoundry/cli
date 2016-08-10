package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type CreateDomainCommand struct {
	RequiredArgs flags.OrgDomain `positional-args:"yes"`
	usage        interface{}     `usage:"CF_NAME create-domain ORG DOMAIN"`
}

func (_ CreateDomainCommand) Setup() error {
	return nil
}

func (_ CreateDomainCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
