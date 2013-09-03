package testhelpers

import (
	"cf/app"
	"flag"
	"github.com/codegangsta/cli"
)

func NewContext(cmdName string, args []string) (*cli.Context) {
	targetCommand := findCommand(cmdName)

	flagSet := new(flag.FlagSet)
	for i, _ := range targetCommand.Flags {
		targetCommand.Flags[i].Apply(flagSet)
	}

	flagSet.Parse(args)

	globalSet := new(flag.FlagSet)

	return cli.NewContext(cli.NewApp(), flagSet, globalSet)
}

func findCommand(cmdName string) (cmd cli.Command) {
	myApp, _ := app.New()

	for _, cmd := range myApp.Commands {
		if cmd.Name == cmdName {
			return cmd
		}
	}

	return
}

