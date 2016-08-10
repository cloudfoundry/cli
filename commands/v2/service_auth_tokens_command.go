package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type ServiceAuthTokensCommand struct {
	usage interface{} `usage:"CF_NAME service-auth-tokens"`
}

func (_ ServiceAuthTokensCommand) Setup() error {
	return nil
}

func (_ ServiceAuthTokensCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
