package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type GetHealthCheckCommand struct {
	RequiredArgs flag.AppName `positional-args:"yes"`
	usage        interface{}  `usage:"CF_NAME get-health-check APP_NAME"`
}

func (_ GetHealthCheckCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ GetHealthCheckCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
