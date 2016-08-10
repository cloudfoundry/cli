package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type StagingEnvironmentVariableGroupCommand struct {
	usage interface{} `usage:"CF_NAME staging-environment-variable-group"`
}

func (_ StagingEnvironmentVariableGroupCommand) Setup() error {
	return nil
}

func (_ StagingEnvironmentVariableGroupCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
