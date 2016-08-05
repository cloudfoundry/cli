package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type UpdateServiceCommand struct {
	RequiredArgs     flags.ServiceInstance `positional-args:"yes"`
	Plan             string                `short:"p" description:"Change service plan for a service instance"`
	ParametersAsJSON string                `short:"c" description:"Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file. For a list of supported configuration parameters, see documentation for the particular service offering."`
	Tags             string                `short:"t" description:"User provided tags"`
}

func (_ UpdateServiceCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
