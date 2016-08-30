package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
)

type StagingEnvironmentVariableGroupCommand struct {
	usage           interface{} `usage:"CF_NAME staging-environment-variable-group"`
	relatedCommands interface{} `related_commands:"env, running-environment-variable-group"`
}

func (_ StagingEnvironmentVariableGroupCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ StagingEnvironmentVariableGroupCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
