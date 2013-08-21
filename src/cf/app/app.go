package app

import (
	"cf/commands"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

func New() (app *cli.App) {
	termUI := new(terminal.TerminalUI)

	app = cli.NewApp()
	app.Name = "cf"
	app.Usage = "A command line tool to interact with Cloud Foundry"
	app.Version = "1.0.0.alpha"
	app.Commands = []cli.Command{
		{
			Name:      "target",
			ShortName: "t",
			Usage:     "Set or view the target",
			Flags: []cli.Flag{
				cli.StringFlag{"o", "", "organization"},
			},
			Action: func(c *cli.Context) {
				commands.Target(c, termUI)
			},
		},
		{
			Name:      "login",
			ShortName: "l",
			Usage:     "Log user in",
			Action: func(c *cli.Context) {
				commands.Login(c, termUI)
			},
		},
	}
	return
}
