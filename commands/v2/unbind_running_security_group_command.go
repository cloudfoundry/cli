package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type UnbindRunningSecurityGroupCommand struct {
	RequiredArgs flags.SecurityGroup `positional-args:"yes"`
	usage        interface{}         `usage:"CF_NAME unbind-running-security-group SECURITY_GROUP\n\nTIP: Changes will not apply to existing running applications until they are restarted."`
}

func (_ UnbindRunningSecurityGroupCommand) Setup() error {
	return nil
}

func (_ UnbindRunningSecurityGroupCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
