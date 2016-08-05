package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type UnbindRouteServiceCommand struct {
	RequiredArgs     flags.RouteServiceArgs `positional-args:"yes"`
	Hostname         string                 `long:"hostname" short:"n" description:"Hostname used in combination with DOMAIN to specify the route to bind"`
	Path             string                 `long:"path" description:"Path for the HTTP route"`
	ParametersAsJSON string                 `short:"c" description:"Valid JSON object containing service-specific configuration parameters, provided inline or in a file. For a list of supported configuration parameters, see documentation for the particular service offering."`
}

func (_ UnbindRouteServiceCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
