package main

import (
	"cf/commands"
	"github.com/codegangsta/cli"
	"os"
	"cf/terminal"
)

func main() {
	app := cli.NewApp()
	app.Name = "cf"
	app.Usage = "A command line tool to interact with Cloud Foundry"
	app.Version = "1.0.0.alpha"
	app.Commands = []cli.Command{
		{
			Name:      "target",
			ShortName: "t",
			Usage:     "Set or view the target",
			Action: func(c *cli.Context) {
				commands.Target(c, new(terminal.ConsoleUI))
			},
		},
		{
			Name:      "login",
			ShortName: "l",
			Usage:     "Log user in",
			Action: func(c *cli.Context) {
				commands.Login(c, new(terminal.ConsoleUI))
			},
		},
	}
	app.Run(os.Args)
}
