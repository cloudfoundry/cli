package commands

import (
	"cf/app"
	"flag"
	"github.com/codegangsta/cli"
	"strings"
	"cf/commands"
	testreq "testhelpers/requirements"
)

func NewContext(cmdName string, args []string) (*cli.Context) {
	targetCommand := findCommand(cmdName)

	flagSet := new(flag.FlagSet)
	for i, _ := range targetCommand.Flags {
		targetCommand.Flags[i].Apply(flagSet)
	}

	// move all flag args to the beginning of the list, go requires them all upfront
	firstFlagIndex := -1
	for index, arg := range args {
		if strings.HasPrefix(arg, "-") {
			firstFlagIndex = index
			break
		}
	}
	if firstFlagIndex > 0 {
		args := args[0:firstFlagIndex]
		flags := args[firstFlagIndex:]
		flagSet.Parse(append(flags, args...))
	} else {
		flagSet.Parse(args[0:])
	}

	globalSet := new(flag.FlagSet)

	return cli.NewContext(cli.NewApp(), flagSet, globalSet)
}

func findCommand(cmdName string) (cmd cli.Command) {
	cmdFactory := commands.ConcreteFactory{}
	reqFactory := &testreq.FakeReqFactory{}
	myApp, _ := app.NewApp(cmdFactory, reqFactory)

	for _, cmd := range myApp.Commands {
		if cmd.Name == cmdName {
			return cmd
		}
	}

	return
}

