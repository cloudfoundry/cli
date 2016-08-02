package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type StacksCommand struct{}

func (_ StacksCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
