package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type CreateServiceKeyCommand struct{}

func (_ CreateServiceKeyCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
