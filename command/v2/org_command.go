package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type OrgCommand struct {
	RequiredArgs    flag.Organization `positional-args:"yes"`
	GUID            bool              `long:"guid" description:"Retrieve and display the given org's guid.  All other output for the org is suppressed."`
	usage           interface{}       `usage:"CF_NAME org ORG [--guid]"`
	relatedCommands interface{}       `related_commands:"org-users, orgs"`
}

func (_ OrgCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ OrgCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
