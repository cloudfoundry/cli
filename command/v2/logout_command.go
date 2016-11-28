package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
)

type LogoutCommand struct {
	usage interface{} `usage:"CF_NAME logout"`
}

func (_ LogoutCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ LogoutCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
