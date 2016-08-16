package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type ScaleCommand struct {
	RequiredArgs flags.AppName `positional-args:"yes"`
	NumInstances int           `short:"i" description:"Number of instances"`
	DiskLimit    string        `short:"k" description:"Disk limit (e.g. 256M, 1024M, 1G)"`
	MemoryLimit  string        `short:"m" description:"Memory limit (e.g. 256M, 1024M, 1G)"`
	ForceRestart bool          `short:"f" description:"Force restart of app without prompt"`
	usage        interface{}   `usage:"CF_NAME scale APP_NAME [-i INSTANCES] [-k DISK] [-m MEMORY] [-f]"`
}

func (_ ScaleCommand) Setup(config commands.Config) error {
	return nil
}

func (_ ScaleCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
