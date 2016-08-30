package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type CreateServiceAuthTokenCommand struct {
	RequiredArgs flags.ServiceAuthTokenArgs `positional-args:"yes"`
	usage        interface{}                `usage:"CF_NAME create-service-auth-token LABEL PROVIDER TOKEN"`
}

func (_ CreateServiceAuthTokenCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ CreateServiceAuthTokenCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
