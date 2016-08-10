package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type UnbindStagingSecurityGroupCommand struct {
	RequiredArgs flags.SecurityGroup `positional-args:"yes"`
	usage        interface{}         `usage:"CF_NAME unbind-staging-security-group SECURITY_GROUP\n\nTIP: Changes will not apply to existing running applications until they are restarted."`
}

func (_ UnbindStagingSecurityGroupCommand) Setup() error {
	return nil
}

func (_ UnbindStagingSecurityGroupCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
