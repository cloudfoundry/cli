package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type DeleteUserCommand struct{}

func (_ DeleteUserCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
