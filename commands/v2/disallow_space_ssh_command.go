package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type DisallowSpaceSSHCommand struct {
	RequiredArgs flags.Space `positional-args:"yes"`
	usage        interface{} `usage:"CF_NAME disallow-space-ssh SPACE_NAME"`
}

func (_ DisallowSpaceSSHCommand) Setup() error {
	return nil
}

func (_ DisallowSpaceSSHCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
