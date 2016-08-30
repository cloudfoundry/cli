package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
)

type AppsCommand struct {
	usage           interface{} `usage:"CF_NAME apps"`
	relatedCommands interface{} `related_commands:"events, logs, map-route, push, scale, start, stop, restart"`
}

func (_ AppsCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ AppsCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
