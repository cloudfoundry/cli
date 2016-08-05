package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type UpdateBuildpackCommand struct {
	RequiredArgs flags.SetSpaceQuotaArgs `positional-args:"yes"`
	Order        int                     `short:"i" description:"The order in which the buildpacks are checked during buildpack auto-detection"`
	Path         int                     `short:"p" description:"Path to directory or zip file"`
	Enable       bool                    `long:"enable" description:"Enable the buildpack to be used for staging"`
	Disable      bool                    `long:"disable" description:"Disable the buildpack from being used for staging"`
	Lock         bool                    `long:"lock" description:"Lock the buildpack to prevent updates"`
	Unlock       bool                    `long:"unlock" description:"Unlock the buildpack to enable updates"`
}

func (_ UpdateBuildpackCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
