package app

import (
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
			Usage:       "cf push --name <application> [--domain <domain>] [--host <hostname>]",
			Flags: []cli.Flag{
				cli.StringFlag{"name", "", "name of the application"},
				cli.StringFlag{"domain", "", "domain (for example: cfapps.io)"},
				cli.StringFlag{"host", "", "hostname (for example: my-subdomain)"},
				cli.IntFlag{"instances", 1, "number of instances"},
			},
			Action: func(c *cli.Context) {
				cmd := commands.NewPush(termUI, config, appRepo, domainRepo, routeRepo)
				cmd.Run(c)
			},
		},
		{
			Name:        "apps",
			ShortName:   "a",
			Description: "List all applications in the currently selected space",
			Usage:       "cf apps",
			Action: func(c *cli.Context) {
				cmd := commands.NewApps(termUI, config, appRepo)
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
	}
	return
}
