package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type SetQuotaCommand struct {
	RequiredArgs flags.SetOrgQuotaArgs `positional-args:"yes"`
	usage        interface{}           `usage:"CF_NAME set-quota ORG QUOTA\n\nTIP:\n    View allowable quotas with 'CF_NAME quotas'"`
}

func (_ SetQuotaCommand) Setup() error {
	return nil
}

func (_ SetQuotaCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
