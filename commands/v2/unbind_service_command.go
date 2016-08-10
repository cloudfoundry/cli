package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type UnbindServiceCommand struct {
	RequiredArgs flags.BindServiceArgs `positional-args:"yes"`
	usage        interface{}           `usage:"CF_NAME unbind-service APP_NAME SERVICE_INSTANCE"`
}

func (_ UnbindServiceCommand) Setup() error {
	return nil
}

func (_ UnbindServiceCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
