package app

import (
	"cf"
	"cf/api"
	"cf/commands"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"fmt"
	"github.com/codegangsta/cli"
	"os"
)

func New() (app *cli.App, err error) {
	cli.AppHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}

USAGE:
   [environment variables] {{.Name}} [global options] command [arguments...] [command options]

VERSION:
   {{.Version}}

COMMANDS:
   {{range .Commands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Description}}
   {{end}}
GLOBAL OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}
ENVIRONMENT VARIABLES:
   CF_TRACE=true - will output HTTP requests and responses during command
   HTTP_PROXY=http://proxy.example.com:8080 - set to your proxy
`

	cli.CommandHelpTemplate = `NAME:
   {{.Name}} - {{.Description}}
{{with .ShortName}}
ALIAS:
   {{.}}
{{end}}
USAGE:
   {{.Usage}}{{with .Flags}}

OPTIONS:
   {{range .}}{{.}}
   {{end}}{{else}}
{{end}}`

	termUI := new(terminal.TerminalUI)
	configRepo := configuration.NewConfigurationDiskRepository()
	config, err := configRepo.Get()
	if err != nil {
		termUI.Failed(fmt.Sprintf(
			"Error loading config. Please reset target (%s) and log in (%s).",
			terminal.CommandColor("cf target"),
			terminal.CommandColor("cf login"),
		))
		configRepo.Delete()
		os.Exit(1)
		return
	}

	repoLocator := api.NewRepositoryLocator(config)
	cmdFactory := commands.NewFactory(termUI, repoLocator)
	reqFactory := requirements.NewFactory(termUI, repoLocator)
	cmdRunner := commands.NewRunner(reqFactory)

	app = cli.NewApp()
	app.Name = "cf"
	app.Usage = "A command line tool to interact with Cloud Foundry"
	app.Version = cf.Version
	app.Commands = []cli.Command{
		{
			Name:        "api",
			Description: "Set or view target api endpoint",
			Usage:       "cf api [URL]",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewApi()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "app",
			Description: "Display health and status for app",
			Usage:       "cf app APP",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewApp()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "apps",
			ShortName:   "a",
			Description: "List all applications in the currently targeted space",
			Usage:       "cf apps",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewApps()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "bind-service",
			ShortName:   "bs",
			Description: "Bind a service instance to an application",
			Usage:       "cf bind-service APP SERVICE_INSTANCE",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewBindService()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "create-org",
			ShortName:   "co",
			Description: "Create organization",
			Usage:       "cf create-org ORG",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewCreateOrganization()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "create-service",
			ShortName:   "cs",
			Description: "Create service instance",
			Usage: "cf create-service --offering [OFFERING] --plan [PLAN] --name SERVICE\n" +
				"   cf create-service --offering user-provided --name SERVICE --parameters \"<comma separated parameter names>\"",
			Flags: []cli.Flag{
				cli.StringFlag{"name", "", "name of the service instance"},
				cli.StringFlag{"offering", "", "name of the service offering to use"},
				cli.StringFlag{"plan", "", "name of the service plan to use"},
				cli.StringFlag{"parameters", "", "list of comma separated parameter names to use for user-provided services (eg. \"n1,n2\")"},
			},
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewCreateService()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "create-space",
			Description: "Create a space",
			Usage:       "cf create-space SPACE",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewCreateSpace()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "delete",
			ShortName:   "d",
			Description: "Delete an application",
			Usage:       "cf delete -f APP",
			Flags: []cli.Flag{
				cli.BoolFlag{"f", "Force deletion without confirmation"},
			},
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewDelete()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "delete-org",
			Description: "Delete an org",
			Usage:       "cf delete-org ORG",
			Flags: []cli.Flag{
				cli.BoolFlag{"f", "force deletion without confirmation"},
			},
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewDeleteOrg()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "delete-service",
			ShortName:   "ds",
			Description: "Delete a service instance",
			Usage:       "cf delete-service SERVICE",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewDeleteService()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "delete-space",
			Description: "Delete a space",
			Usage:       "cf delete-space SPACE",
			Flags: []cli.Flag{
				cli.BoolFlag{"f", "Force deletion without confirmation"},
			},
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewDeleteSpace()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "env",
			ShortName:   "e",
			Description: "Show all env variables for an app",
			Usage:       "cf env APP",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewEnv()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "files",
			ShortName:   "f",
			Description: "Print out a list of files in a directory or the contents of a specific file",
			Usage:       "cf files APP [PATH]",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewFiles()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "login",
			ShortName:   "l",
			Description: "Log user in",
			Usage: "cf login [USERNAME] [PASSWORD]\n\n" +
				terminal.WarningColor("WARNING:\n   Providing your password as a command line option is highly discouraged.\n   Your password may be visible to others and may be recorded in your shell history.\n\n") +
				"EXAMPLE:\n" +
				"   cf login (omit username and password to login interactively -- cf will prompt for both)\n" +
				"   cf login name@example.com pa55woRD (specify username and password to login non-interactively)\n" +
				"   cf login name@example.com \"my password\" (use quotes for passwords with a space)",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewLogin()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "logout",
			ShortName:   "lo",
			Description: "Log user out",
			Usage:       "cf logout",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewLogout()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "logs",
			Description: "Show recent logs for CF applications",
			Usage:       "cf logs APP",
			Flags: []cli.Flag{
				cli.BoolFlag{"recent", "dump recent logs instead of tailing"},
			},
			Action: func(c *cli.Context) {
				var cmd commands.Command
				cmd = cmdFactory.NewLogs()

				if c.Bool("recent") {
					cmd = cmdFactory.NewRecentLogs()
				}

				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "marketplace",
			ShortName:   "m",
			Description: "List available offerings in the marketplace",
			Usage:       "cf marketplace",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewMarketplaceServices()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "org",
			Description: "Show org info",
			Usage:       "cf org ORG",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewShowOrganization()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "orgs",
			ShortName:   "o",
			Description: "List all organizations",
			Usage:       "cf orgs",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewListOrganizations()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "passwd",
			ShortName:   "pw",
			Description: "Change user password",
			Usage:       "cf passwd",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewPassword()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "push",
			ShortName:   "p",
			Description: "Push an app",
			Usage: "cf push --name APP [--domain DOMAIN] [--host HOST] [--instances NUM]\n" +
				"                      [--memory MEMORY] [--buildpack URL] [--no-[re]start] [--path PATH]\n" +
				"                      [--stack STACK]",
			Flags: []cli.Flag{
				cli.StringFlag{"name", "", "name of the application"},
				cli.StringFlag{"domain", "", "domain (for example: cfapps.io)"},
				cli.StringFlag{"host", "", "hostname (for example: my-subdomain)"},
				cli.IntFlag{"instances", 1, "number of instances"},
				cli.StringFlag{"memory", "128", "memory limit (for example: 256, 1G, 1024M)"},
				cli.StringFlag{"buildpack", "", "custom buildpack URL (for example: https://github.com/heroku/heroku-buildpack-play.git)"},
				cli.BoolFlag{"no-start", "do not start an application after pushing"},
				cli.BoolFlag{"no-restart", "do not restart an application after pushing"},
				cli.StringFlag{"path", "", "path of application directory or zip file"},
				cli.StringFlag{"stack", "", "stack to use"},
			},
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewPush()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "rename",
			Description: "Rename an application",
			Usage:       "cf rename APP NEW_APP",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewRename()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "rename-org",
			Description: "Rename an organization",
			Usage:       "cf rename-org ORG NEW_ORG",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewRenameOrg()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "rename-service",
			Description: "Rename a service",
			Usage:       "cf rename-service SERVICE_INSTANCE NEW_SERVICE_INSTANCE",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewRenameService()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "rename-space",
			Description: "Rename a space",
			Usage:       "cf rename-space SPACE NEW_SPACE",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewRenameSpace()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "restart",
			ShortName:   "rs",
			Description: "Restart an application",
			Usage:       "cf restart APP",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewRestart()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "routes",
			ShortName:   "r",
			Description: "List all routes",
			Usage:       "cf routes",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewRoutes()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "scale",
			Description: "Change the disk quota, instance count, and memory limit for an application",
			Usage:       "cf scale APP -d DISK -i INSTANCES -m MEMORY",
			Flags: []cli.Flag{
				cli.StringFlag{"d", "", "disk quota"},
				cli.IntFlag{"i", 0, "number of instances"},
				cli.StringFlag{"m", "", "memory limit"},
			},
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewScale()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "services",
			ShortName:   "s",
			Description: "List all services in the currently targeted space",
			Usage:       "cf services",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewServices()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "set-env",
			ShortName:   "se",
			Description: "Set an env variable for an app",
			Usage:       "cf set-env APP NAME VALUE",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewSetEnv()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "space",
			Description: "Show currently targeted space's info",
			Usage:       "cf space",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewShowSpace()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "spaces",
			Description: "List all spaces in an org",
			Usage:       "cf spaces",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewSpaces()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "stacks",
			Description: "List all stacks",
			Usage:       "cf stacks",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewStacks()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "start",
			ShortName:   "st",
			Description: "Start an app",
			Usage:       "cf start APP",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewStart()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "stop",
			ShortName:   "sp",
			Description: "Stop an app",
			Usage:       "cf stop APP",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewStop()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "target",
			ShortName:   "t",
			Description: "Set or view the targeted org or space",
			Usage:       "cf target [-o ORG] [-s SPACE]",
			Flags: []cli.Flag{
				cli.StringFlag{"o", "", "organization"},
				cli.StringFlag{"s", "", "space"},
			},
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewTarget()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "unbind-service",
			ShortName:   "us",
			Description: "Unbind a service instance from an application",
			Usage:       "cf unbind-service APP SERVICE_INSTANCE",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewUnbindService()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "unset-env",
			Description: "Remove an env variable",
			Usage:       "cf unset-env APP NAME",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewUnsetEnv()
				cmdRunner.Run(cmd, c)
			},
		},
	}
	return
}
