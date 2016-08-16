package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
)

type DomainsCommand struct {
	usage interface{} `usage:"CF_NAME domains"`
}

func (_ DomainsCommand) Setup(config commands.Config) error {
	return nil
}

func (_ DomainsCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
