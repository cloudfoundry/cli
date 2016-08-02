package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type HelpCommand struct{}

func (_ HelpCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
