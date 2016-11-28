package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
)

type StagingSecurityGroupsCommand struct {
	usage           interface{} `usage:"CF_NAME staging-security-groups"`
	relatedCommands interface{} `related_commands:"bind-staging-security-group, security-group, unbind-staging-security-group"`
}

func (_ StagingSecurityGroupsCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ StagingSecurityGroupsCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
