package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type StacksCommand struct {
	usage interface{} `usage:"CF_NAME stacks"`
}

func (_ StacksCommand) Setup() error {
	return nil
}

func (_ StacksCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
