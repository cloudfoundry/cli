package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type UnsetEnvCommand struct {
	usage interface{} `usage:"CF_NAME unset-env APP_NAME ENV_VAR_NAME"`
}

func (_ UnsetEnvCommand) Setup() error {
	return nil
}

func (_ UnsetEnvCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
