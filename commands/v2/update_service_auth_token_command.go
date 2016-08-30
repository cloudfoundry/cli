package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type UpdateServiceAuthTokenCommand struct {
	RequiredArgs flags.ServiceAuthTokenArgs `positional-args:"yes"`
	usage        interface{}                `usage:"CF_NAME update-service-auth-token LABEL PROVIDER TOKEN"`
}

func (_ UpdateServiceAuthTokenCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ UpdateServiceAuthTokenCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
