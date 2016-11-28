package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
)

type AppsCommand struct {
	usage           interface{} `usage:"CF_NAME apps"`
	relatedCommands interface{} `related_commands:"events, logs, map-route, push, scale, start, stop, restart"`
}

func (_ AppsCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ AppsCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
