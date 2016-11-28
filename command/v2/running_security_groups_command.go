package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
)

type RunningSecurityGroupsCommand struct {
	usage           interface{} `usage:"CF_NAME running-security-groups"`
	relatedCommands interface{} `related_commands:"bind-running-security-group, security-group, unbind-running-security-group"`
}

func (_ RunningSecurityGroupsCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ RunningSecurityGroupsCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
