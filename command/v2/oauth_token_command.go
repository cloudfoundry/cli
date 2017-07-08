package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
)

type OauthTokenCommand struct {
	usage           interface{} `usage:"CF_NAME oauth-token"`
	relatedCommands interface{} `related_commands:"curl"`
}

func (OauthTokenCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (OauthTokenCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
