package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
)

type SSHCodeCommand struct {
	usage           interface{} `usage:"CF_NAME ssh-code"`
	relatedCommands interface{} `related_commands:"curl, ssh"`
}

func (_ SSHCodeCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ SSHCodeCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
