package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type SpaceCommand struct {
	RequiredArgs       flags.Space `positional-args:"yes"`
	GUID               bool        `long:"guid" description:"Retrieve and display the given space's guid.  All other output for the space is suppressed."`
	SecurityGroupRules bool        `long:"security-group-rules" description:"Retrieve the rules for all the security groups associated with the space"`
}

func (_ SpaceCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
