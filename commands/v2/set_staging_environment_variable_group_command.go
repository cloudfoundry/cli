package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type SetStagingEnvironmentVariableGroupCommand struct {
	RequiredArgs flags.ParamsAsJSON `positional-args:"yes"`
}

func (_ SetStagingEnvironmentVariableGroupCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
