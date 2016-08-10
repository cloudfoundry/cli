package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type AllowSpaceSSHCommand struct {
	RequiredArgs flags.Space `positional-args:"yes"`
	usage        interface{} `usage:"CF_NAME allow-space-ssh SPACE_NAME"`
}

func (_ AllowSpaceSSHCommand) Setup() error {
	return nil
}

func (_ AllowSpaceSSHCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
