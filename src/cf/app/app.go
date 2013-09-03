package app

import (
	"cf"
	"cf/api"
	"cf/commands"
	"cf/configuration"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

func New() (app *cli.App) {
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
	config := configuration.Get()

	organizationRepo := new(api.CloudControllerOrganizationRepository)
	spaceRepo := new(api.CloudControllerSpaceRepository)
	appRepo := new(api.CloudControllerApplicationRepository)
	domainRepo := new(api.CloudControllerDomainRepository)
	routeRepo := new(api.CloudControllerRouteRepository)
	stackRepo := new(api.CloudControllerStackRepository)
	serviceRepo := new(api.CloudControllerServiceRepository)

	app = cli.NewApp()
	app.Name = "cf"
	app.Usage = "A command line tool to interact with Cloud Foundry"
	app.Version = "1.0.0.alpha"
	app.Commands = []cli.Command{
		{
			Name:        "target",
			ShortName:   "t",
			Description: "Set or view the target",
			Usage:       "cf target <target> --o <organization> --s <space>",
			Flags: []cli.Flag{
				cli.StringFlag{"o", "", "organization"},
				cli.StringFlag{"s", "", "space"},
			},
			Action: func(c *cli.Context) {
				cmd := commands.NewTarget(termUI, config, organizationRepo, spaceRepo)
				cmd.Run(c)
			},
		},
		{
			Name:        "login",
			ShortName:   "l",
			Description: "Log user in",
			Usage:       "cf login [username]",
			Action: func(c *cli.Context) {
				authenticator := new(api.UAAAuthenticator)

				cmd := commands.NewLogin(termUI, config, organizationRepo, spaceRepo, authenticator)
				cmd.Run(c)
			},
		},
		{
			Name:        "set-env",
			ShortName:   "se",
			Description: "Set an environment variable for an application",
			Usage:       "cf set-env <application> <variable> <value>",
			Action: func(c *cli.Context) {
				cmd := commands.NewSetEnv(termUI, config, appRepo)
				cmd.Run(c)
			},
		},
		{
			Name:        "logout",
			ShortName:   "lo",
			Description: "Log user out",
			Usage:       "cf logout",
			Action: func(c *cli.Context) {
				cmd := commands.NewLogout(termUI, config)
				cmd.Run(c)
			},
		},
		{
			Name:        "push",
			ShortName:   "p",
			Description: "Push an application",
			Usage: "cf push --name <application> [--domain <domain>] [--host <hostname>] [--instances <num>]\n" +
				"                                [--memory <memory>] [--buildpack <url>] [--no-start] [--path <path to app>]\n" +
				"                                [--stack <stack>]",
			Flags: []cli.Flag{
				cli.StringFlag{"name", "", "name of the application"},
				cli.StringFlag{"domain", "", "domain (for example: cfapps.io)"},
				cli.StringFlag{"host", "", "hostname (for example: my-subdomain)"},
				cli.IntFlag{"instances", 1, "number of instances"},
				cli.StringFlag{"memory", "128", "memory limit (for example: 256, 1G, 1024M)"},
				cli.StringFlag{"buildpack", "", "custom buildpack URL (for example: https://github.com/heroku/heroku-buildpack-play.git)"},
				cli.BoolFlag{"no-start", "do not start an application after pushing"},
				cli.StringFlag{"path", "", "path of application directory or zip file"},
				cli.StringFlag{"stack", "", "stack to use"},
			},
			Action: func(c *cli.Context) {
				startCmd := commands.NewStart(termUI, config, appRepo)
				zipper := cf.ApplicationZipper{}
				cmd := commands.NewPush(termUI, config, &startCmd, zipper, appRepo, domainRepo, routeRepo, stackRepo)
				cmd.Run(c)
			},
		},
		{
			Name:        "apps",
			ShortName:   "a",
			Description: "List all applications in the currently selected space",
			Usage:       "cf apps",
			Action: func(c *cli.Context) {
				cmd := commands.NewApps(termUI, config, spaceRepo)
				cmd.Run(c)
			},
		},
		{
			Name:        "delete",
			ShortName:   "d",
			Description: "Delete an application",
			Usage:       "cf delete <application>",
			Action: func(c *cli.Context) {
				cmd := commands.NewDelete(termUI, config, appRepo)
				cmd.Run(c)
			},
		},
		{
			Name:        "start",
			ShortName:   "s",
			Description: "Start applications",
			Usage:       "cf start <application>",
			Action: func(c *cli.Context) {
				cmd := commands.NewStart(termUI, config, appRepo)
				cmd.Run(c)
			},
		},
		{
			Name:        "stop",
			ShortName:   "st",
			Description: "Stop applications",
			Usage:       "cf stop <application>",
			Action: func(c *cli.Context) {
				cmd := commands.NewStop(termUI, config, appRepo)
				cmd.Run(c)
			},
		},
		{
			Name:        "create-service",
			ShortName:   "cs",
			Description: "Create service instance",
			Usage:       "cf create-service --offering <offering> --plan <plan> --name <service instance name>",
			Flags: []cli.Flag{
				cli.StringFlag{"name", "", "name of the service instance"},
				cli.StringFlag{"offering", "", "name of the service offering to use"},
				cli.StringFlag{"plan", "", "name of the service plan to use"},
			},
			Action: func(c *cli.Context) {
				cmd := commands.NewCreateService(termUI, config, serviceRepo)
				cmd.Run(c)
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
				cmd := commands.NewBindService(termUI, config, serviceRepo, appRepo)
				cmd.Run(c)
			},
		},
	}
	return
}
