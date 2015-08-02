package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/command_runner"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/cf/trace"
	"github.com/codegangsta/cli"
)

const UnknownCommand = "cf: '%s' is not a registered command. See 'cf help'"

func NewApp(cmdRunner command_runner.Runner, metadatas ...command_metadata.CommandMetadata) (app *cli.App) {
	helpCommand := cli.Command{
		Name:        "help",
		ShortName:   "h",
		Description: T("Show help"),
		Usage:       fmt.Sprintf(T("{{.Command}} help [COMMAND]", map[string]interface{}{"Command": cf.Name()})),
		Action: func(c *cli.Context) {
			args := c.Args()
			if len(args) > 0 {
				cli.ShowCommandHelp(c, args[0])
			}
		},
	}

	trace.Logger.Printf("\n%s\n%s\n\n", terminal.HeaderColor(T("VERSION:")), cf.Version)

	app = cli.NewApp()
	app.Version = cf.Version + "-" + cf.BuiltOnDate
	app.Action = helpCommand.Action
	app.CommandNotFound = func(c *cli.Context, command string) {
		panic(errors.Exception{
			Message:            fmt.Sprintf(UnknownCommand, command),
			DisplayCrashDialog: false,
		})
	}

	compiledAtTime, err := time.Parse("2006-01-02T03:04:05+00:00", cf.BuiltOnDate)

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
			err := runner.RunCmdByName(metadata.Name, context)
			if err != nil {
				panic(terminal.QuietPanic)
			}
		},
		Flags:           metadata.Flags,
		SkipFlagParsing: metadata.SkipFlagParsing,
	}
}
