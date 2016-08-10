package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type OrgCommand struct {
	RequiredArgs flags.Organization `positional-args:"yes"`
	GUID         bool               `long:"guid" description:"Retrieve and display the given org's guid.  All other output for the org is suppressed."`
	usage        interface{}        `usage:"CF_NAME org ORG"`
}

func (_ OrgCommand) Setup() error {
	return nil
}

func (_ OrgCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
