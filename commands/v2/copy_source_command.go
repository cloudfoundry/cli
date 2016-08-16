package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type CopySourceCommand struct {
	RequiredArgs flags.CopySourceArgs `positional-args:"yes"`
	NoRestart    bool                 `long:"no-restart" description:"Override restart of the application in target environment after copy-source completes"`
	Organization string               `short:"o" description:"Org that contains the target application"`
	Space        string               `short:"s" description:"Space that contains the target application"`
	usage        interface{}          `usage:"CF_NAME copy-source SOURCE-APP TARGET-APP [-s TARGET-SPACE [-o TARGET-ORG]] [--no-restart]\n"`
}

func (_ CopySourceCommand) Setup(config commands.Config) error {
	return nil
}

func (_ CopySourceCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
