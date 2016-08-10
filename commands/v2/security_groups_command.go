package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type SecurityGroupsCommand struct {
	usage interface{} `usage:"CF_NAME security-groups"`
}

func (_ SecurityGroupsCommand) Setup() error {
	return nil
}

func (_ SecurityGroupsCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
