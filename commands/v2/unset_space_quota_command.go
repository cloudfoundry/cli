package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type UnsetSpaceQuotaCommand struct {
	RequiredArgs flags.SetSpaceQuotaArgs `positional-args:"yes"`
	usage        interface{}             `usage:"CF_NAME unset-space-quota SPACE QUOTA\n\n"`
}

func (_ UnsetSpaceQuotaCommand) Setup() error {
	return nil
}

func (_ UnsetSpaceQuotaCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
