package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type OrgCommand struct {
	RequiredArgs    flags.Organization `positional-args:"yes"`
	GUID            bool               `long:"guid" description:"Retrieve and display the given org's guid.  All other output for the org is suppressed."`
	usage           interface{}        `usage:"CF_NAME org ORG [--guid]"`
	relatedCommands interface{}        `related_commands:"org-users, orgs"`
}

func (_ OrgCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ OrgCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
