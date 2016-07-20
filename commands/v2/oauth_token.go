package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type OauthTokenCommand struct{}

func (_ OauthTokenCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
