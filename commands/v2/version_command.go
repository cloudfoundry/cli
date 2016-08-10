package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type VersionCommand struct {
	usage interface{} `usage:"CF_NAME version\n\n    \"cf -v\" and \"cf --version\" are also accepted."`
}

func (_ VersionCommand) Setup() error {
	return nil
}

func (_ VersionCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
