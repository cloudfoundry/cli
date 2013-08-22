package testhelpers

import (
	"cf/app"
	"flag"
	"github.com/codegangsta/cli"
)

func NewContext(cmdIndex int, args []string) (*cli.Context) {
	app := app.New()
	targetCommand := app.Commands[cmdIndex]

	flagSet := new(flag.FlagSet)
	for i, _ := range targetCommand.Flags {
		targetCommand.Flags[i].Apply(flagSet)
	}

	flagSet.Parse(args)

	globalSet := new(flag.FlagSet)

	return cli.NewContext(cli.NewApp(), flagSet, globalSet)
}
