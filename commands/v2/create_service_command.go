package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type CreateServiceCommand struct {
	RequiredArgs      flags.CreateServiceArgs `positional-args:"yes"`
	ConfigurationFile string                  `short:"c" description:"Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file. For a list of supported configuration parameters, see documentation for the particular service offering."`
	Tags              string                  `short:"t" description:"User provided tags"`
}

func (_ CreateServiceCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
