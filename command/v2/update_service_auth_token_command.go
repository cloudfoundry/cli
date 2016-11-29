package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type UpdateServiceAuthTokenCommand struct {
	RequiredArgs flag.ServiceAuthTokenArgs `positional-args:"yes"`
	usage        interface{}               `usage:"CF_NAME update-service-auth-token LABEL PROVIDER TOKEN"`
}

func (_ UpdateServiceAuthTokenCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ UpdateServiceAuthTokenCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
