package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type CreateBuildpackCommand struct {
	RequiredArgs flags.CreateBuildpackArgs `positional-args:"yes"`
	Enable       bool                      `long:"enable" description:"Enable the buildpack to be used for staging"`
	Disable      bool                      `long:"disable" description:"Disable the buildpack from being used for staging"`
}

func (_ CreateBuildpackCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
