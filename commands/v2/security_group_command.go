package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type SecurityGroupCommand struct {
	RequiredArgs flags.SecurityGroup `positional-args:"yes"`
	usage        interface{}         `usage:"CF_NAME security-group SECURITY_GROUP"`
}

func (_ SecurityGroupCommand) Setup(config commands.Config) error {
	return nil
}

func (_ SecurityGroupCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
