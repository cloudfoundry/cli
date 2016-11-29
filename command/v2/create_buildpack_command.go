package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type CreateBuildpackCommand struct {
	RequiredArgs    flag.CreateBuildpackArgs `positional-args:"yes"`
	Disable         bool                     `long:"disable" description:"Disable the buildpack from being used for staging"`
	Enable          bool                     `long:"enable" description:"Enable the buildpack to be used for staging"`
	usage           interface{}              `usage:"CF_NAME create-buildpack BUILDPACK PATH POSITION [--enable|--disable]\n\nTIP:\n   Path should be a zip file, a url to a zip file, or a local directory. Position is a positive integer, sets priority, and is sorted from lowest to highest."`
	relatedCommands interface{}              `related_commands:"buildpacks, push"`
}

func (_ CreateBuildpackCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (c CreateBuildpackCommand) Execute(args []string) error {
	_, err := flag.ParseStringToInt(c.RequiredArgs.Position)
	if err != nil {
		return command.ParseArgumentError{
			ArgumentName: "POSITION",
			ExpectedType: "integer",
		}
	}

	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
