package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type UnbindServiceCommand struct {
	RequiredArgs flags.BindServiceArgs `positional-args:"yes"`
}

func (_ UnbindServiceCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
