package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type AppsCommand struct {
	usage interface{} `usage:"CF_NAME apps"`
}

func (_ AppsCommand) Setup() error {
	return nil
}

func (_ AppsCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
