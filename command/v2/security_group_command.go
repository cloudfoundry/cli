package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type SecurityGroupCommand struct {
	RequiredArgs    flag.SecurityGroup `positional-args:"yes"`
	usage           interface{}        `usage:"CF_NAME security-group SECURITY_GROUP"`
	relatedCommands interface{}        `related_commands:"bind-security-group, bind-running-security-group, bind-staging-security-group"`
}

func (_ SecurityGroupCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ SecurityGroupCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
