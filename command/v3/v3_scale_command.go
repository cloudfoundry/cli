package v3

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type V3ScaleCommand struct {
	RequiredArgs    flag.AppName `positional-args:"yes"`
	NumInstances    int          `short:"i" description:"Number of instances"`
	DiskLimit       string       `short:"k" description:"Disk limit (e.g. 256M, 1024M, 1G)"`
	MemoryLimit     string       `short:"m" description:"Memory limit (e.g. 256M, 1024M, 1G)"`
	usage           interface{}  `usage:"CF_NAME v3-scale APP_NAME [-i INSTANCES] [-k DISK] [-m MEMORY]"`
	relatedCommands interface{}  `related_commands:"v3-push"`
}

func (V3ScaleCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (V3ScaleCommand) Execute(args []string) error {
	return nil
}
