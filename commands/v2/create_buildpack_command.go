package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type CreateBuildpackCommand struct {
	RequiredArgs    flags.CreateBuildpackArgs `positional-args:"yes"`
	Disable         bool                      `long:"disable" description:"Disable the buildpack from being used for staging"`
	Enable          bool                      `long:"enable" description:"Enable the buildpack to be used for staging"`
	usage           interface{}               `usage:"CF_NAME create-buildpack BUILDPACK PATH POSITION [--enable|--disable]\n\nTIP:\n   Path should be a zip file, a url to a zip file, or a local directory. Position is a positive integer, sets priority, and is sorted from lowest to highest."`
	relatedCommands interface{}               `related_commands:"buildpacks, push"`
}

func (_ CreateBuildpackCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ CreateBuildpackCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
