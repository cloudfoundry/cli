package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type SetStagingEnvironmentVariableGroupCommand struct {
	RequiredArgs    flag.ParamsAsJSON `positional-args:"yes"`
	usage           interface{}       `usage:"CF_NAME set-staging-environment-variable-group '{\"name\":\"value\",\"name\":\"value\"}'"`
	relatedCommands interface{}       `related_commands:"set-env, staging-environment-variable-group"`
}

func (_ SetStagingEnvironmentVariableGroupCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ SetStagingEnvironmentVariableGroupCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
