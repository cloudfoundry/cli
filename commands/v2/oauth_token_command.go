package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type OauthTokenCommand struct {
	usage interface{} `usage:"CF_NAME oauth-token"`
}

func (_ OauthTokenCommand) Setup() error {
	return nil
}

func (_ OauthTokenCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
