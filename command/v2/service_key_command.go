package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type ServiceKeyCommand struct {
	RequiredArgs flag.ServiceInstanceKey `positional-args:"yes"`
	GUID         bool                    `long:"guid" description:"Retrieve and display the given service-key's guid.  All other output for the service is suppressed."`
	usage        interface{}             `usage:"CF_NAME service-key SERVICE_INSTANCE SERVICE_KEY\n\nEXAMPLES:\n   CF_NAME service-key mydb mykey"`
}

func (_ ServiceKeyCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ ServiceKeyCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
