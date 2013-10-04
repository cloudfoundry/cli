package app

import (
	"cf"
	"cf/commands"
	"cf/requirements"
	"cf/terminal"
	"fmt"
	"github.com/codegangsta/cli"
)

func NewApp(cmdFactory commands.Factory, reqFactory requirements.Factory) (app *cli.App, err error) {

	cmdRunner := commands.NewRunner(reqFactory)

	app = cli.NewApp()
	app.Name = cf.Name
	app.Usage = cf.Usage
	app.Version = cf.Version
	app.Commands = []cli.Command{
		{
			Name:        "api",
			Description: "Set or view target api url",
			Usage:       fmt.Sprintf("%s api [URL]", cf.Name),
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("api")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "app",
			Description: "Display health and status for app",
			Usage:       fmt.Sprintf("%s app APP", cf.Name),
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("app")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "apps",
			ShortName:   "a",
			Description: "List all applications in the currently targeted space",
			Usage:       fmt.Sprintf("%s apps", cf.Name),
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("apps")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "bind-service",
			ShortName:   "bs",
			Description: "Bind a service instance to an application",
			Usage:       fmt.Sprintf("%s bind-service APP SERVICE_INSTANCE", cf.Name),
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("bind-service")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "create-org",
			ShortName:   "co",
			Description: "Create organization",
			Usage:       fmt.Sprintf("%s create-org ORG", cf.Name),
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("create-org")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "create-service",
			ShortName:   "cs",
			Description: "Create service instance",
			Usage: fmt.Sprintf("%s create-service SERVICE PLAN SERVICE_INSTANCE\n\n", cf.Name) +
				"EXAMPLE:\n" +
				fmt.Sprintf("   %s create-service clear-db spark clear-db-mine\n\n", cf.Name) +
				"TIP:\n" +
				"   Use 'cf create-user-provided-service' to make user-provided services available to cf apps",
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("create-service")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "create-space",
			Description: "Create a space",
			Usage:       fmt.Sprintf("%s create-space SPACE", cf.Name),
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("create-space")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "create-user-provided-service",
			ShortName:   "cups",
			Description: "Make a user-provided service available to cf apps",
			Usage: fmt.Sprintf("%s create-service SERVICE_INSTANCE \"comma, separated, parameter, names\"\n\n", cf.Name) +
				"EXAMPLE:\n" +
				fmt.Sprintf("   %s create-user-provided-service oracle-db-mine \"host, port, dbname, username, password\"", cf.Name),
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("create-user-provided-service")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "delete",
			ShortName:   "d",
			Description: "Delete an application",
			Usage:       fmt.Sprintf("%s delete -f APP", cf.Name),
			Flags: []cli.Flag{
				cli.BoolFlag{"f", "Force deletion without confirmation"},
			},
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("delete")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "delete-org",
			Description: "Delete an org",
			Usage:       fmt.Sprintf("%s delete-org ORG", cf.Name),
			Flags: []cli.Flag{
				cli.BoolFlag{"f", "force deletion without confirmation"},
			},
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("delete-org")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "delete-service",
			ShortName:   "ds",
			Description: "Delete a service instance",
			Usage:       fmt.Sprintf("%s delete-service SERVICE", cf.Name),
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("delete-service")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "delete-space",
			Description: "Delete a space",
			Usage:       fmt.Sprintf("%s delete-space SPACE", cf.Name),
			Flags: []cli.Flag{
				cli.BoolFlag{"f", "Force deletion without confirmation"},
			},
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("delete-space")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "domains",
			Description: "List domains in the currently targeted org",
			Usage:       fmt.Sprintf("%s domains", cf.Name),
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("domains")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "env",
			ShortName:   "e",
			Description: "Show all env variables for an app",
			Usage:       fmt.Sprintf("%s env APP", cf.Name),
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("env")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "files",
			ShortName:   "f",
			Description: "Print out a list of files in a directory or the contents of a specific file",
			Usage:       fmt.Sprintf("%s files APP [PATH]", cf.Name),
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("files")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "login",
			ShortName:   "l",
			Description: "Log user in",
			Usage: fmt.Sprintf("%s login [USERNAME] [PASSWORD]\n\n", cf.Name) +
				terminal.WarningColor("WARNING:\n   Providing your password as a command line option is highly discouraged.\n   Your password may be visible to others and may be recorded in your shell history.\n\n") +
				"EXAMPLE:\n" +
				fmt.Sprintf("   %s login (omit username and password to login interactively -- %s will prompt for both)\n", cf.Name, cf.Name) +
				fmt.Sprintf("   %s login name@example.com pa55woRD (specify username and password to login non-interactively)\n", cf.Name) +
				fmt.Sprintf("   %s login name@example.com \"my password\" (use quotes for passwords with a space)", cf.Name),
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("login")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "logout",
			ShortName:   "lo",
			Description: "Log user out",
			Usage:       fmt.Sprintf("%s logout", cf.Name),
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("logout")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "logs",
			Description: "Show recent logs for applications",
			Usage:       fmt.Sprintf("%s logs APP", cf.Name),
			Flags: []cli.Flag{
				cli.BoolFlag{"recent", "dump recent logs instead of tailing"},
			},
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("logs")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "marketplace",
			ShortName:   "m",
			Description: "List available offerings in the marketplace",
			Usage:       fmt.Sprintf("%s marketplace", cf.Name),
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("marketplace")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "map-domain",
			Description: "Map a domain to a space",
			Usage:       fmt.Sprintf("%s map-domain SPACE DOMAIN", cf.Name),
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("map-domain")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "map-route",
			Description: "Add a url route to an app",
			Usage:       fmt.Sprintf("%s map-route APP DOMAIN [-n HOSTNAME]", cf.Name),
			Flags: []cli.Flag{
				cli.StringFlag{"n", "", "Hostname"},
			},
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("map-route")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "org",
			Description: "Show org info",
			Usage:       fmt.Sprintf("%s org ORG", cf.Name),
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("org")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "orgs",
			ShortName:   "o",
			Description: "List all organizations",
			Usage:       fmt.Sprintf("%s orgs", cf.Name),
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("orgs")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "passwd",
			ShortName:   "pw",
			Description: "Change user password",
			Usage:       fmt.Sprintf("%s passwd", cf.Name),
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("passwd")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "push",
			ShortName:   "p",
			Description: "Push an app",
			Usage: fmt.Sprintf("%s push APP [-d DOMAIN] [-n HOST] [-i NUM_INSTANCES]\n", cf.Name) +
				"               [-m MEMORY] [-b URL] [--no-[re]start] [-p PATH]\n" +
				"               [-s STACK] [-c COMMAND]",
			Flags: []cli.Flag{
				cli.StringFlag{"d", "", "domain (for example: example.com)"},
				cli.StringFlag{"n", "", "hostname (for example: my-subdomain)"},
				cli.IntFlag{"i", 1, "number of instances"},
				cli.StringFlag{"m", "128", "memory limit (for example: 256, 1G, 1024M)"},
				cli.StringFlag{"b", "", "custom buildpack URL (for example: https://github.com/heroku/heroku-buildpack-play.git)"},
				cli.BoolFlag{"no-start", "do not start an application after pushing"},
				cli.BoolFlag{"no-restart", "do not restart an application after pushing"},
				cli.StringFlag{"p", "", "path of application directory or zip file"},
				cli.StringFlag{"s", "", "stack to use"},
				cli.StringFlag{"c", "", "startup command"},
			},
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("push")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "rename",
			Description: "Rename an application",
			Usage:       fmt.Sprintf("%s rename APP NEW_APP", cf.Name),
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("rename")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "rename-org",
			Description: "Rename an organization",
			Usage:       fmt.Sprintf("%s rename-org ORG NEW_ORG", cf.Name),
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("rename-org")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "rename-service",
			Description: "Rename a service instance",
			Usage:       fmt.Sprintf("%s rename-service SERVICE_INSTANCE NEW_SERVICE_INSTANCE", cf.Name),
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("rename-service")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "rename-space",
			Description: "Rename a space",
			Usage:       fmt.Sprintf("%s rename-space SPACE NEW_SPACE", cf.Name),
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("rename-space")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "reserve-domain",
			Description: "Add a domain to an org",
			Usage:       fmt.Sprintf("%s reserve-domain ORG DOMAIN", cf.Name),
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("reserve-domain")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "reserve-route",
			Description: "Reserve a url route on a space for later use",
			Usage:       fmt.Sprintf("%s reserve-route SPACE DOMAIN [-n HOSTNAME]", cf.Name),
			Flags: []cli.Flag{
				cli.StringFlag{"n", "", "Hostname"},
			},
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("reserve-route")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "restart",
			ShortName:   "rs",
			Description: "Restart an application",
			Usage:       fmt.Sprintf("%s restart APP", cf.Name),
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("restart")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "routes",
			ShortName:   "r",
			Description: "List all routes",
			Usage:       fmt.Sprintf("%s routes", cf.Name),
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("routes")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "scale",
			Description: "Change the disk quota, instance count, and memory limit for an application",
			Usage:       fmt.Sprintf("%s scale APP -d DISK -i INSTANCES -m MEMORY", cf.Name),
			Flags: []cli.Flag{
				cli.StringFlag{"d", "", "disk quota"},
				cli.IntFlag{"i", 0, "number of instances"},
				cli.StringFlag{"m", "", "memory limit"},
			},
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("scale")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "service",
			Description: "Show service instance info",
			Usage:       fmt.Sprintf("%s service SERVICE_INSTANCE", cf.Name),
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("service")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "services",
			ShortName:   "s",
			Description: "List all services in the currently targeted space",
			Usage:       fmt.Sprintf("%s services", cf.Name),
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("services")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "set-env",
			ShortName:   "se",
			Description: "Set an env variable for an app",
			Usage:       fmt.Sprintf("%s set-env APP NAME VALUE", cf.Name),
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("set-env")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "set-quota",
			Description: "Define the quota for an org",
			Usage: fmt.Sprintf("%s set-quota ORG QUOTA\n\n", cf.Name) +
				"TIP:\n" +
				"   Allowable quotas are 'free,' 'paid,' 'runaway,' and 'trial'",
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("set-quota")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "space",
			Description: "Show currently targeted space's info",
			Usage:       fmt.Sprintf("%s space", cf.Name),
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("space")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "spaces",
			Description: "List all spaces in an org",
			Usage:       fmt.Sprintf("%s spaces", cf.Name),
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("spaces")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "stacks",
			Description: "List all stacks",
			Usage:       fmt.Sprintf("%s stacks", cf.Name),
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("stacks")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "start",
			ShortName:   "st",
			Description: "Start an app",
			Usage:       fmt.Sprintf("%s start APP", cf.Name),
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("start")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "stop",
			ShortName:   "sp",
			Description: "Stop an app",
			Usage:       fmt.Sprintf("%s stop APP", cf.Name),
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("stop")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "target",
			ShortName:   "t",
			Description: "Set or view the targeted org or space",
			Usage:       fmt.Sprintf("%s target [-o ORG] [-s SPACE]", cf.Name),
			Flags: []cli.Flag{
				cli.StringFlag{"o", "", "organization"},
				cli.StringFlag{"s", "", "space"},
			},
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("target")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "unbind-service",
			ShortName:   "us",
			Description: "Unbind a service instance from an application",
			Usage:       fmt.Sprintf("%s unbind-service APP SERVICE_INSTANCE", cf.Name),
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("unbind-service")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "unmap-route",
			Description: "Remove a url route from an app",
			Usage:       fmt.Sprintf("%s unmap-route APP DOMAIN [-n HOSTNAME]", cf.Name),
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("unmap-route")
				cmdRunner.Run(cmd, c)
			},
		},
		{
			Name:        "unset-env",
			Description: "Remove an env variable",
			Usage:       fmt.Sprintf("%s unset-env cf.Name", cf.Name),
			Action: func(c *cli.Context) {
				cmd, _ := cmdFactory.GetByCmdName("unset-env")
				cmdRunner.Run(cmd, c)
			},
		},
	}
	return
}
