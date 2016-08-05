package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type CreateRouteCommand struct {
	RequiredArgs flags.SpaceDomain `positional-args:"yes"`
	Hostname     string            `long:"hostname" short:"n" description:"Hostname for the HTTP route (required for shared domains)"`
	Path         string            `long:"path" description:"Path for the HTTP route"`
	Port         int               `long:"port" description:"Port for the TCP route"`
	RandomPort   bool              `long:"random-port" description:"Create a random port for the TCP route"`
}

func (_ CreateRouteCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
