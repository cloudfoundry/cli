package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type CreateUserProvidedServiceCommand struct {
	RequiredArgs    flags.ServiceInstance `positional-args:"yes"`
	Credentials     string                `short:"p" description:"Credentials, provided inline or in a file, to be exposed in the VCAP_SERVICES environment variable for bound applications"`
	SyslogDrainURL  string                `short:"l" description:"URL to which logs for bound applications will be streamed"`
	RouteServiceURL string                `short:"r" description:"URL to which requests for bound routes will be forwarded. Scheme for this URL must be https"`
}

func (_ CreateUserProvidedServiceCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
