package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
)

type QuotasCommand struct {
	usage interface{} `usage:"CF_NAME quotas"`
}

func (QuotasCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (QuotasCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
