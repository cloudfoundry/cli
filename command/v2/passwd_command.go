package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
)

type PasswdCommand struct {
	usage interface{} `usage:"CF_NAME passwd"`
}

func (_ PasswdCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ PasswdCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
