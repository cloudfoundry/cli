package main

import (
	"cf/configuration"
	term "cf/terminal"
	"github.com/codegangsta/cli"
	"os"
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
					config := configuration.Default()

					term.Say("Target Information (where will apps be pushed):")

					term.Say("CF instance: %s (API version: %s)",
						term.Yellow(config.Target),
						term.Yellow(config.ApiVersion))

					term.Say("user: %s",
						term.Red("N/A"))

					term.Say("target app space: %s (org: %s)",
						term.Red("N/A"),
						term.Red("N/A"))

					return
			},
		},
	}
	app.Run(os.Args)
}
