package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
)

type RunningEnvironmentVariableGroupCommand struct {
	usage           interface{} `usage:"CF_NAME running-environment-variable-group"`
	relatedCommands interface{} `related_commands:"env, staging-environment-variable-group"`
}

func (_ RunningEnvironmentVariableGroupCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ RunningEnvironmentVariableGroupCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
