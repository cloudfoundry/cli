package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type UnmapRouteCommand struct {
	RequiredArgs flags.AppDomain `positional-args:"yes"`
	Hostname     string          `long:"hostname" short:"n" description:"Hostname used to identify the HTTP route"`
	Path         string          `long:"path" description:"Path used to identify the HTTP route"`
	Port         int             `long:"port" description:"Port used to identify the TCP route"`
}

func (_ UnmapRouteCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
