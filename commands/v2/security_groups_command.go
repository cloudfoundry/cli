package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
)

type SecurityGroupsCommand struct {
	usage interface{} `usage:"CF_NAME security-groups"`
}

func (_ SecurityGroupsCommand) Setup(config commands.Config) error {
	return nil
}

func (_ SecurityGroupsCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
