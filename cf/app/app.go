package app

import (
	"fmt"
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/command_runner"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/cf/trace"
	"github.com/codegangsta/cli"
	"strings"
	"time"
)

var appHelpTemplate = `{{.Title "NAME:"}}
   {{.Name}} - {{.Usage}}

{{.Title "USAGE:"}}
   [environment variables] {{.Name}} [global options] command [arguments...] [command options]

{{.Title "VERSION:"}}
   {{.Version}}

{{.Title "BUILD TIME:"}}
   {{.Compiled}}
   {{range .Commands}}
{{.SubTitle .Name}}{{range .CommandSubGroups}}
{{range .}}   {{.Name}} {{.Description}}
{{end}}{{end}}{{end}}
{{.Title "ENVIRONMENT VARIABLES"}}
   CF_COLOR=false                     Do not colorize output
   CF_HOME=path/to/dir/               Override path to default config directory
   CF_STAGING_TIMEOUT=15              Max wait time for buildpack staging, in minutes
   CF_STARTUP_TIMEOUT=5               Max wait time for app instance startup, in minutes
   CF_TRACE=true                      Print API request diagnostics to stdout
   CF_TRACE=path/to/trace.log         Append API request diagnostics to a log file
   HTTP_PROXY=proxy.example.com:8080  Enable HTTP proxying for API requests

{{.Title "GLOBAL OPTIONS"}}
   --version, -v                      Print the version
   --help, -h                         Show help
`

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
		Flags:           metadata.Flags,
		SkipFlagParsing: metadata.SkipFlagParsing,
	}
}
