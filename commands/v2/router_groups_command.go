package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type RouterGroupsCommand struct {
	usage interface{} `usage:"CF_NAME router-groups"`
}

func (_ RouterGroupsCommand) Setup() error {
	return nil
}

func (_ RouterGroupsCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
