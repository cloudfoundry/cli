package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type DeleteBuildpackCommand struct {
	RequiredArgs flags.Buildpack `positional-args:"yes"`
	Force        bool            `short:"f" description:"Force deletion without confirmation"`
	usage        interface{}     `usage:"CF_NAME delete-buildpack BUILDPACK [-f]"`
}

func (_ DeleteBuildpackCommand) Setup() error {
	return nil
}

func (_ DeleteBuildpackCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
