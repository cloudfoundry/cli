package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type CopySourceCommand struct{}

func (_ CopySourceCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
