package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type DeleteServiceAuthTokenCommand struct {
	RequiredArgs flag.DeleteServiceAuthTokenArgs `positional-args:"yes"`
	Force        bool                            `short:"f" description:"Force deletion without confirmation"`
	usage        interface{}                     `usage:"CF_NAME delete-service-auth-token LABEL PROVIDER [-f]"`
}

func (_ DeleteServiceAuthTokenCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ DeleteServiceAuthTokenCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
