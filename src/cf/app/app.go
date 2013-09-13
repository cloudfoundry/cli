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
   [environment variables] {{.Name}} [global options] command [command options] [arguments...]

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
			Usage:       "cf api [api endpoint]",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewApi()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "apps",
			ShortName:   "a",
			Description: "List all applications in the currently selected space",
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
			Usage:       "cf bind-service --app <application name> --service <service instance name>",
			Flags: []cli.Flag{
				cli.StringFlag{"app", "", "name of the application"},
				cli.StringFlag{"service", "", "name of the service instance to bind to the application"},
			},
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewBindService()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "create-org",
			ShortName:   "co",
			Description: "Create organization",
			Usage:       "cf create-org <organization name>",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewCreateOrganization()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "create-service",
			ShortName:   "cs",
			Description: "Create service instance",
			Usage: "cf create-service --offering <offering> --plan <plan> --name <service instance name>\n" +
				"   cf create-service --offering user-provided --name <service name> --parameters \"<comma separated parameter names>\"",
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
			Usage:       "cf create-space spacename",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewCreateSpace()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "delete",
			ShortName:   "d",
			Description: "Delete an application",
			Usage:       "cf delete -f <application>",
			Flags: []cli.Flag{
				cli.BoolFlag{"f", "force deletion without confirmation"},
			},
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewDelete()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "delete-org",
			Description: "Delete an org",
			Usage:       "cf delete-org <org-name>",
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
			Usage:       "cf delete-service <service instance name>",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewDeleteService()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "delete-space",
			Description: "Delete a space",
			Usage:       "cf delete-space <space-name>",
			Flags: []cli.Flag{
				cli.BoolFlag{"f", "force deletion without confirmation"},
			},
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewDeleteSpace()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "files",
			Description: "Print out a list of files in a directory or the contents of a specific file",
			Usage:       "cf files <appname> [file or directory]",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewFiles()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "login",
			ShortName:   "l",
			Description: "Log user in",
			Usage:       "cf login [username [password]]\n\n" + terminal.WarningColor("WARNING:\nProviding your password as a command line option is highly discouraged.\nYour password may be visible to others and may be recorded in your shell history."),
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
			Description: "Push an application",
			Usage: "cf push --name <application> [--domain <domain>] [--host <hostname>] [--instances <num>]\n" +
				"                                [--memory <memory>] [--buildpack <url>] [--no-[re]start] [--path <path to app>]\n" +
				"                                [--stack <stack>]",
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
			Name:        "rename-org",
			Description: "Rename an organization",
			Usage:       "cf rename-space <org-name> <new-org-name>",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewRenameOrg()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "rename-space",
			Description: "Rename a space",
			Usage:       "cf rename-space <current-space-name> <new-space-name>",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewRenameSpace()
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
			Name:        "services",
			ShortName:   "s",
			Description: "List all services in the currently selected space",
			Usage:       "cf services [--marketplace]",
			Flags: []cli.Flag{
				cli.BoolFlag{"marketplace", "use to list available offerings on the marketplace"},
			},
			Action: func(c *cli.Context) {
				var cmd commands.Command
				cmd = cmdFactory.NewServices()

				if c.Bool("marketplace") {
					cmd = cmdFactory.NewMarketplaceServices()
				}

				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "set-env",
			ShortName:   "se",
			Description: "Set an environment variable for an application",
			Usage:       "cf set-env <application> <variable> <value>",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewSetEnv()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "spaces",
			Description: "List all spaces in an organization",
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
			Description: "Start applications",
			Usage:       "cf start <application>",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewStart()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "stop",
			ShortName:   "sp",
			Description: "Stop applications",
			Usage:       "cf stop <application>",
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewStop()
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "target",
			ShortName:   "t",
			Description: "Set or view the target",
			Usage:       "cf target [--o <organization>] [--s <space>]",
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
			Usage:       "cf unbind-service --app <application name> --service <service instance name>",
			Flags: []cli.Flag{
				cli.StringFlag{"app", "", "name of the application"},
				cli.StringFlag{"service", "", "name of the service instance to unbind from the application"},
			},
			Action: func(c *cli.Context) {
				cmd := cmdFactory.NewUnbindService()
				cmdRunner.Run(cmd, c)
			},
		},
	}
	return
}
