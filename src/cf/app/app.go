package app

import (
	"cf"
	"cf/command_metadata"
	"cf/command_runner"
	"cf/terminal"
	"cf/trace"
	"fmt"
	"github.com/codegangsta/cli"
	"strings"
	"time"
)

func NewApp(cmdRunner command_runner.Runner, metadatas ...command_metadata.CommandMetadata) (app *cli.App) {
	helpCommand := cli.Command{
		Name:        "help",
		ShortName:   "h",
		Description: "Show help",
		Usage:       fmt.Sprintf("%s help [COMMAND]", cf.Name()),
		Action: func(c *cli.Context) {
			args := c.Args()
			if len(args) > 0 {
				cli.ShowCommandHelp(c, args[0])
			} else {
				showAppHelp(appHelpTemplate, c.App)
			}
		},
	}
	cli.HelpPrinter = showAppHelp
	cli.AppHelpTemplate = appHelpTemplate

	trace.Logger.Printf("\n%s\n%s\n\n", terminal.HeaderColor("VERSION:"), cf.Version)

	app = cli.NewApp()
	app.Usage = cf.Usage
	app.Version = cf.Version
	app.Action = helpCommand.Action

	compiledAtTime, err := time.Parse("Jan 2, 2006 3:04PM", cf.BuiltOnDate)

	if err == nil {
		app.Compiled = compiledAtTime
	} else {
		err = nil
		app.Compiled = time.Now()
	}

	app.Commands = []cli.Command{helpCommand}

	for _, metadata := range metadatas {
		app.Commands = append(app.Commands, getCommand(metadata, cmdRunner))
	}
	return
}

func getCommand(metadata command_metadata.CommandMetadata, runner command_runner.Runner) cli.Command {
	return cli.Command{
		Name:        metadata.Name,
		ShortName:   metadata.ShortName,
		Description: metadata.Description,
		Usage:       strings.Replace(metadata.Usage, "CF_NAME", cf.Name(), -1),
		Action: func(context *cli.Context) {
			runner.RunCmdByName(metadata.Name, context)
		},
		Flags: metadata.Flags,
	}
}
