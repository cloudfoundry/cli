package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type OrgsCommand struct{}

func (_ OrgsCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
