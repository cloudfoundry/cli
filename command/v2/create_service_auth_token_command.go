package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type CreateServiceAuthTokenCommand struct {
	RequiredArgs flag.ServiceAuthTokenArgs `positional-args:"yes"`
	usage        interface{}               `usage:"CF_NAME create-service-auth-token LABEL PROVIDER TOKEN"`
}

func (_ CreateServiceAuthTokenCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ CreateServiceAuthTokenCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
