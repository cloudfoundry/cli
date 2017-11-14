package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type ScaleCommand struct {
	RequiredArgs    flag.AppName `positional-args:"yes"`
	ForceRestart    bool         `short:"f" description:"Force restart of app without prompt"`
	NumInstances    int          `short:"i" description:"Number of instances"`
	DiskLimit       string       `short:"k" description:"Disk limit (e.g. 256M, 1024M, 1G)"`
	MemoryLimit     string       `short:"m" description:"Memory limit (e.g. 256M, 1024M, 1G)"`
	usage           interface{}  `usage:"CF_NAME scale APP_NAME [-i INSTANCES] [-k DISK] [-m MEMORY] [-f]"`
	relatedCommands interface{}  `related_commands:"push"`
}

func (ScaleCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (ScaleCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
