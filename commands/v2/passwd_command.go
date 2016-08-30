package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
)

type PasswdCommand struct {
	usage interface{} `usage:"CF_NAME passwd"`
}

func (_ PasswdCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ PasswdCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
