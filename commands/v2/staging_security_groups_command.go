package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type StagingSecurityGroupsCommand struct {
	usage interface{} `usage:"CF_NAME staging-security-groups"`
}

func (_ StagingSecurityGroupsCommand) Setup() error {
	return nil
}

func (_ StagingSecurityGroupsCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
