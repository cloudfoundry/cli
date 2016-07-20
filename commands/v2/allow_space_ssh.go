package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type AllowSpaceSSHCommand struct{}

func (_ AllowSpaceSSHCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
