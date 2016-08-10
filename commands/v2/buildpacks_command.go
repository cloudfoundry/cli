package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type BuildpacksCommand struct {
	usage interface{} `usage:"CF_NAME buildpacks"`
}

func (_ BuildpacksCommand) Setup() error {
	return nil
}

func (_ BuildpacksCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
