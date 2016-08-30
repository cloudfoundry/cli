package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type ServiceKeyCommand struct {
	RequiredArgs flags.ServiceInstanceKey `positional-args:"yes"`
	GUID         bool                     `long:"guid" description:"Retrieve and display the given service-key's guid.  All other output for the service is suppressed."`
	usage        interface{}              `usage:"CF_NAME service-key SERVICE_INSTANCE SERVICE_KEY\n\nEXAMPLES:\n   CF_NAME service-key mydb mykey"`
}

func (_ ServiceKeyCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ ServiceKeyCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
