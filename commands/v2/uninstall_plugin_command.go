package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type UninstallPluginCommand struct{}

func (_ UninstallPluginCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
