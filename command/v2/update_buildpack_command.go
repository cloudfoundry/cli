package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type UpdateBuildpackCommand struct {
	RequiredArgs    flag.Buildpack `positional-args:"yes"`
	Disable         bool           `long:"disable" description:"Disable the buildpack from being used for staging"`
	Enable          bool           `long:"enable" description:"Enable the buildpack to be used for staging"`
	Order           int            `short:"i" description:"The order in which the buildpacks are checked during buildpack auto-detection"`
	Lock            bool           `long:"lock" description:"Lock the buildpack to prevent updates"`
	Path            string         `short:"p" description:"Path to directory or zip file"`
	Unlock          bool           `long:"unlock" description:"Unlock the buildpack to enable updates"`
	usage           interface{}    `usage:"CF_NAME update-buildpack BUILDPACK [-p PATH] [-i POSITION] [--enable|--disable] [--lock|--unlock]\n\nTIP:\n   Path should be a zip file, a url to a zip file, or a local directory. Position is a positive integer, sets priority, and is sorted from lowest to highest."`
	relatedCommands interface{}    `related_commands:"buildpacks, rename-buildpack"`
}

func (_ UpdateBuildpackCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ UpdateBuildpackCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
