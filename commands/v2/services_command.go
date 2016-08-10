package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type ServicesCommand struct {
	usage interface{} `usage:"CF_NAME services"`
}

func (_ ServicesCommand) Setup() error {
	return nil
}

func (_ ServicesCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
