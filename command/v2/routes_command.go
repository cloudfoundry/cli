package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
)

type RoutesCommand struct {
	OrgLevel        bool        `long:"orglevel" description:"List all the routes for all spaces of current organization"`
	usage           interface{} `usage:"CF_NAME routes [--orglevel]"`
	relatedCommands interface{} `related_commands:"check-route, domains, map-route, unmap-route"`
}

func (_ RoutesCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ RoutesCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
