package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type SetStagingEnvironmentVariableGroupCommand struct{}

func (_ SetStagingEnvironmentVariableGroupCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
