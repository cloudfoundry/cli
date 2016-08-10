package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type SSHCodeCommand struct {
	usage interface{} `usage:"CF_NAME ssh-code"`
}

func (_ SSHCodeCommand) Setup() error {
	return nil
}

func (_ SSHCodeCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
