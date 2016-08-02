package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type RenameSpaceCommand struct{}

func (_ RenameSpaceCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
