package main

import (
	"cf/configuration"
	termcolor "cf/terminalcolor"
	"github.com/codegangsta/cli"
	"os"
	"fmt"
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

				fmt.Println("Target Information (where will apps be pushed):")
				fmt.Printf("CF instance: %s (API version: %s)\n",
					termcolor.Colorize(config.Target, termcolor.Yellow, true),
					termcolor.Colorize(config.ApiVersion, termcolor.Cyan, true))
				fmt.Println("user: N/A")
				fmt.Println("target app space: N/A (org: N/A)")
			},
		},
	}
	app.Run(os.Args)
}
