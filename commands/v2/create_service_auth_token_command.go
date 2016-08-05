package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type CreateServiceAuthTokenCommand struct {
	RequiredArgs flags.ServiceAuthTokenArgs `positional-args:"yes"`
}

func (_ CreateServiceAuthTokenCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
