package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type RunningEnvironmentVariableGroupCommand struct {
	usage interface{} `usage:"CF_NAME running-environment-variable-group"`
}

func (_ RunningEnvironmentVariableGroupCommand) Setup() error {
	return nil
}

func (_ RunningEnvironmentVariableGroupCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
