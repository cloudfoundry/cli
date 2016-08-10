package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type PluginsCommand struct {
	Checksum bool        `long:"checksum" description:"Compute and show the sha1 value of the plugin binary file"`
	usage    interface{} `usage:"CF_NAME plugins"`
}

func (_ PluginsCommand) Setup() error {
	return nil
}

func (_ PluginsCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
