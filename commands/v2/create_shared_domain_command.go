package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type CreateSharedDomainCommand struct {
	RequiredArgs flags.Domain `positional-args:"yes"`
	RouterGroup  string       `long:"router-group" description:"Routes for this domain will be configured only on the specified router group"`
}

func (_ CreateSharedDomainCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
