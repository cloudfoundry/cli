package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
)

type OauthTokenCommand struct {
	usage           interface{} `usage:"CF_NAME oauth-token"`
	relatedCommands interface{} `related_commands:"curl"`
}

func (_ OauthTokenCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ OauthTokenCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
