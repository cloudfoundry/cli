package v2

import (
	"fmt"
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type AppCommand struct {
	RequiredArgs flags.AppName `positional-args:"yes"`
	GUID         bool          `long:"guid" description:"Retrieve and display the given app's guid.  All other health and status output for the app is suppressed."`
}

func (_ AppCommand) Execute(args []string) error {
	fmt.Println("executing app command")
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
