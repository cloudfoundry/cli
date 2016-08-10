package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type DomainsCommand struct {
	usage interface{} `usage:"CF_NAME domains"`
}

func (_ DomainsCommand) Setup() error {
	return nil
}

func (_ DomainsCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
