package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
)

type SecurityGroupsCommand struct {
	usage           interface{} `usage:"CF_NAME security-groups"`
	relatedCommands interface{} `related_commands:"bind-security-group, bind-running-security-group, bind-staging-security-group, security-group"`
}

func (_ SecurityGroupsCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ SecurityGroupsCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
