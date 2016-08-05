package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type DeleteRouteCommand struct {
	RequiredArgs flags.Domain `positional-args:"yes"`
	Force        bool         `short:"f" description:"Force deletion without confirmation"`
	Hostname     string       `long:"hostname" short:"n" description:"Hostname used to identify the HTTP route"`
	Path         string       `long:"path" description:"Path used to identify the HTTP route"`
	Port         int          `long:"port" description:"Port used to identify the TCP route"`
}

func (_ DeleteRouteCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
