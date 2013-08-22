package app

import (
	"cf/api"
	"cf/commands"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

func New() (app *cli.App) {
	termUI := new(terminal.TerminalUI)
	organizationRepo := new(api.CloudControllerOrganizationRepository)
	spaceRepo := new(api.CloudControllerSpaceRepository)

	app = cli.NewApp()
	app.Name = "cf"
	app.Usage = "A command line tool to interact with Cloud Foundry"
	app.Version = "1.0.0.alpha"
	app.Commands = []cli.Command{
		{
			Name:        "target",
			ShortName:   "t",
			Description: "Set or view the target",
			Flags: []cli.Flag{
				cli.StringFlag{"o", "", "organization"},
				cli.StringFlag{"s", "", "space"},
			},
			Action: func(c *cli.Context) {
				cmd := commands.NewTarget(termUI, organizationRepo, spaceRepo)
				cmd.Run(c)
			},
		},
		{
			Name:        "login",
			ShortName:   "l",
			Description: "Log user in",
			Action: func(c *cli.Context) {
				cmd := commands.NewLogin(termUI, organizationRepo, spaceRepo)
				cmd.Run(c)
			},
		},
		{
			Name:        "set-env",
			ShortName:   "se",
			Description: "Set an environment variable for an application",
			Action: func(c *cli.Context) {
				appRepo := new(api.CloudControllerApplicationRepository)

				cmd := commands.NewSetEnv(termUI, appRepo)
				cmd.Run(c)
			},
		},
	}
	return
}
