package app

import (
	"cf"
	"cf/commands"
	"cf/terminal"
	"cf/trace"
	"fmt"
	"github.com/codegangsta/cli"
)

func NewApp(cmdRunner commands.Runner) (app *cli.App, err error) {
	helpCommand := cli.Command{
		Name:        "help",
		ShortName:   "h",
		Description: "Show help",
		Usage:       fmt.Sprintf("%s help [COMMAND]", cf.Name()),
		Action: func(c *cli.Context) {
			args := c.Args()
			if len(args) > 0 {
				cli.ShowCommandHelp(c, args[0])
			} else {
				showAppHelp(appHelpTemplate, c.App)
			}
		},
	}
	cli.HelpPrinter = showAppHelp
	cli.AppHelpTemplate = appHelpTemplate

	trace.Logger.Printf("\n%s\n%s\n\n", terminal.HeaderColor("VERSION:"), cf.Version)

	app = cli.NewApp()
	app.Usage = cf.Usage
	app.Version = cf.Version
	app.Action = helpCommand.Action
	app.Commands = []cli.Command{
		helpCommand,
		{
			Name:        "api",
			Description: "Set or view target api url",
			Usage:       fmt.Sprintf("%s api [URL]", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("api", c)
			},
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "skip-ssl-validation", Usage: "Please don't"},
			},
		},
		{
			Name:        "app",
			Description: "Display health and status for app",
			Usage:       fmt.Sprintf("%s app APP", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("app", c)
			},
		},
		{
			Name:        "apps",
			ShortName:   "a",
			Description: "List all apps in the target space",
			Usage:       fmt.Sprintf("%s apps", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("apps", c)
			},
		},
		{
			Name:        "auth",
			Description: "Authenticate user non-interactively",
			Usage: fmt.Sprintf("%s auth USERNAME PASSWORD\n\n", cf.Name()) +
				terminal.WarningColor("WARNING:\n   Providing your password as a command line option is highly discouraged\n   Your password may be visible to others and may be recorded in your shell history\n\n") +
				"EXAMPLE:\n" +
				fmt.Sprintf("   %s auth name@example.com \"my password\" (use quotes for passwords with a space)\n", cf.Name()) +
				fmt.Sprintf("   %s auth name@example.com \"\\\"password\\\"\" (escape quotes if used in password)", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("auth", c)
			},
		},
		{
			Name:        "bind-service",
			ShortName:   "bs",
			Description: "Bind a service instance to an app",
			Usage:       fmt.Sprintf("%s bind-service APP SERVICE_INSTANCE", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("bind-service", c)
			},
		},
		{
			Name:        "buildpacks",
			Description: "List all buildpacks",
			Usage:       fmt.Sprintf("%s buildpacks", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("buildpacks", c)
			},
		},
		{
			Name:        "create-buildpack",
			Description: "Create a buildpack",
			Usage: fmt.Sprintf("%s create-buildpack BUILDPACK PATH POSITION [--enable|--disable]", cf.Name()) +
				"\n\nTIP:\n" +
				"   Path should be a zip file, a url to a zip file, or a local directory. Position is an integer, sets priority, and is sorted from lowest to highest.",
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "enable", Usage: "Enable the buildpack"},
				cli.BoolFlag{Name: "disable", Usage: "Disable the buildpack"},
			},
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("create-buildpack", c)
			},
		},
		{
			Name:        "create-domain",
			Description: "Create a domain in an org for later use",
			Usage:       fmt.Sprintf("%s create-domain ORG DOMAIN", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("create-domain", c)
			},
		},
		{
			Name:        "create-org",
			ShortName:   "co",
			Description: "Create an org",
			Usage:       fmt.Sprintf("%s create-org ORG", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("create-org", c)
			},
		},
		{
			Name:        "create-route",
			Description: "Create a url route in a space for later use",
			Usage:       fmt.Sprintf("%s create-route SPACE DOMAIN [-n HOSTNAME]", cf.Name()),
			Flags: []cli.Flag{
				NewStringFlag("n", "Hostname"),
			},
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("create-route", c)
			},
		},
		{
			Name:        "create-service",
			ShortName:   "cs",
			Description: "Create a service instance",
			Usage: fmt.Sprintf("%s create-service SERVICE PLAN SERVICE_INSTANCE\n\n", cf.Name()) +
				"EXAMPLE:\n" +
				fmt.Sprintf("   %s create-service cleardb spark clear-db-mine\n\n", cf.Name()) +
				"TIP:\n" +
				"   Use '" + cf.Name() + " create-user-provided-service' to make user-provided services available to cf apps",
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("create-service", c)
			},
		},
		{
			Name:        "create-service-auth-token",
			Description: "Create a service auth token",
			Usage:       fmt.Sprintf("%s create-service-auth-token LABEL PROVIDER TOKEN", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("create-service-auth-token", c)
			},
		},
		{
			Name:        "create-service-broker",
			Description: "Create a service broker",
			Usage:       fmt.Sprintf("%s create-service-broker SERVICE_BROKER USERNAME PASSWORD URL", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("create-service-broker", c)
			},
		},
		{
			Name:        "create-space",
			Description: "Create a space",
			Usage:       fmt.Sprintf("%s create-space SPACE [-o ORG]", cf.Name()),
			Flags: []cli.Flag{
				NewStringFlag("o", "Organization"),
			},
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("create-space", c)
			},
		},
		{
			Name:        "create-user",
			Description: "Create a new user",
			Usage:       fmt.Sprintf("%s create-user USERNAME PASSWORD", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("create-user", c)
			},
		},
		{
			Name:        "create-user-provided-service",
			ShortName:   "cups",
			Description: "Make a user-provided service available to cf apps",
			Usage: fmt.Sprintf("%s create-user-provided-service SERVICE_INSTANCE [-p PARAMETERS] [-l SYSLOG-DRAIN-URL]\n", cf.Name()) +
				"\n   Pass comma separated parameter names to enable interactive mode:\n" +
				fmt.Sprintf("   %s create-user-provided-service SERVICE_INSTANCE -p \"comma, separated, parameter, names\"\n", cf.Name()) +
				"\n   Pass parameters as JSON to create a service non-interactively:\n" +
				fmt.Sprintf("   %s create-user-provided-service SERVICE_INSTANCE -p '{\"name\":\"value\",\"name\":\"value\"}'\n", cf.Name()) +
				"\nEXAMPLE:\n" +
				fmt.Sprintf("   %s create-user-provided-service oracle-db-mine -p \"host, port, dbname, username, password\"\n", cf.Name()) +
				fmt.Sprintf("   %s create-user-provided-service oracle-db-mine -p '{\"username\":\"admin\",\"password\":\"pa55woRD\"}'\n", cf.Name()) +
				fmt.Sprintf("   %s create-user-provided-service my-drain-service -l syslog://example.com\n", cf.Name()),
			Flags: []cli.Flag{
				NewStringFlag("p", "Parameters"),
				NewStringFlag("l", "Syslog Drain Url"),
			},
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("create-user-provided-service", c)
			},
		},
		{
			Name:        "curl",
			Description: "Executes a raw request, content-type set to application/json by default",
			Usage:       fmt.Sprintf("%s curl PATH [-X METHOD] [-H HEADER] [-d DATA] [-i]", cf.Name()),
			Flags: []cli.Flag{
				cli.StringFlag{Name: "X", Value: "GET", Usage: "HTTP method (GET,POST,PUT,DELETE,etc)"},
				NewStringSliceFlag("H", "Custom headers to include in the request, flag can be specified multiple times"),
				NewStringFlag("d", "HTTP data to include in the request body"),
				cli.BoolFlag{Name: "i", Usage: "Include response headers in the output"},
				cli.BoolFlag{Name: "v", Usage: "Enable CF_TRACE output for all requests and responses"},
			},
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("curl", c)
			},
		},
		{
			Name:        "delete",
			ShortName:   "d",
			Description: "Delete an app",
			Usage:       fmt.Sprintf("%s delete APP [-f -r]", cf.Name()),
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "f", Usage: "Force deletion without confirmation"},
				cli.BoolFlag{Name: "r", Usage: "Also delete any mapped routes"},
			},
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("delete", c)
			},
		},
		{
			Name:        "delete-buildpack",
			Description: "Delete a buildpack",
			Usage:       fmt.Sprintf("%s delete-buildpack BUILDPACK [-f]", cf.Name()),
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "f", Usage: "Force deletion without confirmation"},
			},
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("delete-buildpack", c)
			},
		},
		{
			Name:        "delete-domain",
			Description: "Delete a domain",
			Usage:       fmt.Sprintf("%s delete-domain DOMAIN [-f]", cf.Name()),
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "f", Usage: "Force deletion without confirmation"},
			},
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("delete-domain", c)
			},
		},
		{
			Name:        "delete-shared-domain",
			Description: "Delete a shared domain",
			Usage:       fmt.Sprintf("%s delete-shared-domain DOMAIN [-f]", cf.Name()),
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "f", Usage: "Force deletion without confirmation"},
			},
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("delete-shared-domain", c)
			},
		},
		{
			Name:        "delete-org",
			Description: "Delete an org",
			Usage:       fmt.Sprintf("%s delete-org ORG [-f]", cf.Name()),
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "f", Usage: "Force deletion without confirmation"},
			},
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("delete-org", c)
			},
		},
		{
			Name:        "delete-orphaned-routes",
			Description: "Delete all orphaned routes",
			Usage:       fmt.Sprintf("%s delete-orphaned-routes [-f]", cf.Name()),
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "f", Usage: "Force deletion without confirmation"},
			},
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("delete-orphaned-routes", c)
			},
		},
		{
			Name:        "delete-route",
			Description: "Delete a route",
			Usage:       fmt.Sprintf("%s delete-route DOMAIN [-n HOSTNAME] [-f]", cf.Name()),
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "f", Usage: "Force deletion without confirmation"},
				NewStringFlag("n", "Hostname"),
			},
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("delete-route", c)
			},
		},
		{
			Name:        "delete-service",
			ShortName:   "ds",
			Description: "Delete a service instance",
			Usage:       fmt.Sprintf("%s delete-service SERVICE_INSTANCE [-f]", cf.Name()),
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "f", Usage: "Force deletion without confirmation"},
			},
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("delete-service", c)
			},
		},
		{
			Name:        "delete-service-auth-token",
			Description: "Delete a service auth token",
			Usage:       fmt.Sprintf("%s delete-service-auth-token LABEL PROVIDER [-f]", cf.Name()),
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "f", Usage: "Force deletion without confirmation"},
			},
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("delete-service-auth-token", c)
			},
		},
		{
			Name:        "delete-service-broker",
			Description: "Delete a service broker",
			Usage:       fmt.Sprintf("%s delete-service-broker SERVICE_BROKER [-f]", cf.Name()),
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "f", Usage: "Force deletion without confirmation"},
			},
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("delete-service-broker", c)
			},
		},
		{
			Name:        "delete-space",
			Description: "Delete a space",
			Usage:       fmt.Sprintf("%s delete-space SPACE [-f]", cf.Name()),
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "f", Usage: "Force deletion without confirmation"},
			},
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("delete-space", c)
			},
		},
		{
			Name:        "delete-user",
			Description: "Delete a user",
			Usage:       fmt.Sprintf("%s delete-user USERNAME [-f]", cf.Name()),
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "f", Usage: "Force deletion without confirmation"},
			},
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("delete-user", c)
			},
		},
		{
			Name:        "domains",
			Description: "List domains in the target org",
			Usage:       fmt.Sprintf("%s domains", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("domains", c)
			},
		},
		{
			Name:        "env",
			ShortName:   "e",
			Description: "Show all env variables for an app",
			Usage:       fmt.Sprintf("%s env APP", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("env", c)
			},
		},
		{
			Name:        "events",
			Description: "Show recent app events",
			Usage:       fmt.Sprintf("%s events APP", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("events", c)
			},
		},
		{
			Name:        "files",
			ShortName:   "f",
			Description: "Print out a list of files in a directory or the contents of a specific file",
			Usage:       fmt.Sprintf("%s files APP [PATH]", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("files", c)
			},
		},
		{
			Name:        "login",
			ShortName:   "l",
			Description: "Log user in",
			Usage: fmt.Sprintf("%s login [-a API_URL] [-u USERNAME] [-p PASSWORD] [-o ORG] [-s SPACE]\n\n", cf.Name()) +
				terminal.WarningColor("WARNING:\n   Providing your password as a command line option is highly discouraged\n   Your password may be visible to others and may be recorded in your shell history\n\n") +
				"EXAMPLE:\n" +
				fmt.Sprintf("   %s login (omit username and password to login interactively -- %s will prompt for both)\n", cf.Name(), cf.Name()) +
				fmt.Sprintf("   %s login -u name@example.com -p pa55woRD (specify username and password as arguments)\n", cf.Name()) +
				fmt.Sprintf("   %s login -u name@example.com -p \"my password\" (use quotes for passwords with a space)\n", cf.Name()) +
				fmt.Sprintf("   %s login -u name@example.com -p \"\\\"password\\\"\" (escape quotes if used in password)", cf.Name()),
			Flags: []cli.Flag{
				StringFlagWithNoDefault{cli.StringFlag{
					Name: "a", Usage: "API endpoint (e.g. https://api.example.com)",
				}},
				NewStringFlag("u", "Username"),
				NewStringFlag("p", "Password"),
				NewStringFlag("o", "Org"),
				NewStringFlag("s", "Space"),
				cli.BoolFlag{Name: "skip-ssl-validation", Usage: "Please don't"},
			},
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("login", c)
			},
		},
		{
			Name:        "logout",
			ShortName:   "lo",
			Description: "Log user out",
			Usage:       fmt.Sprintf("%s logout", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("logout", c)
			},
		},
		{
			Name:        "logs",
			Description: "Tail or show recent logs for an app",
			Usage:       fmt.Sprintf("%s logs APP", cf.Name()),
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "recent", Usage: "Dump recent logs instead of tailing"},
			},

			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("logs", c)
			},
		},
		{
			Name:        "marketplace",
			ShortName:   "m",
			Description: "List available offerings in the marketplace",
			Usage:       fmt.Sprintf("%s marketplace", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("marketplace", c)
			},
		},
		{
			Name:        "map-route",
			Description: "Add a url route to an app",
			Usage:       fmt.Sprintf("%s map-route APP DOMAIN [-n HOSTNAME]", cf.Name()),
			Flags: []cli.Flag{
				NewStringFlag("n", "Hostname"),
			},
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("map-route", c)
			},
		},
		{
			Name:        "org",
			Description: "Show org info",
			Usage:       fmt.Sprintf("%s org ORG", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("org", c)
			},
		},
		{
			Name:        "org-users",
			Description: "Show org users by role",
			Usage:       fmt.Sprintf("%s org-users ORG", cf.Name()),
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "a", Usage: "List all users in the org"},
			},
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("org-users", c)
			},
		},
		{
			Name:        "orgs",
			ShortName:   "o",
			Description: "List all orgs",
			Usage:       fmt.Sprintf("%s orgs", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("orgs", c)
			},
		},
		{
			Name:        "passwd",
			ShortName:   "pw",
			Description: "Change user password",
			Usage:       fmt.Sprintf("%s passwd", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("passwd", c)
			},
		},
		{
			Name:        "purge-service-offering",
			Description: "Recursively remove a service and child objects from Cloud Foundry database without making requests to a service broker",
			Usage: fmt.Sprintf("%s purge-service-offering SERVICE [-p PROVIDER]", cf.Name()) +
				"\n\nWARNING:\n" +
				"This operation assumes that the service broker responsible for this service offering is no longer available, and all service instances have been deleted, leaving orphan records in Cloud Foundry's database. All knowledge of the service will be removed from Cloud Foundry, including service instances and service bindings. No attempt will be made to contact the service broker; running this command without destroying the service broker will cause orphan service instances. After running this command you may want to run either delete-service-auth-token or delete-service-broker to complete the cleanup.",
			Flags: []cli.Flag{
				NewStringFlag("p", "Provider"),
				cli.BoolFlag{Name: "f", Usage: "Force deletion without confirmation"},
			},
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("purge-service-offering", c)
			},
		},
		{
			Name:        "push",
			ShortName:   "p",
			Description: "Push a new app or sync changes to an existing app",
			Usage: "Push a single app (with or without a manifest):\n" +
				fmt.Sprintf("   %s push APP [-b BUILDPACK_NAME] [-c COMMAND] [-d DOMAIN] [-f MANIFEST_PATH]\n", cf.Name()) +
				"   [-i NUM_INSTANCES] [-k DISK] [-m MEMORY] [-n HOST] [-p PATH] [-s STACK] [-t TIMEOUT]\n" +
				"   [--no-hostname] [--no-manifest] [--no-route] [--no-start]" +
				"\n\n   Push multiple apps with a manifest:\n" +
				fmt.Sprintf("   %s push [-f MANIFEST_PATH]\n", cf.Name()),
			Flags: []cli.Flag{
				NewStringFlag("b", "Custom buildpack by name (e.g. my-buildpack) or GIT URL (e.g. https://github.com/heroku/heroku-buildpack-play.git)"),
				NewStringFlag("c", "Startup command, set to null to reset to default start command"),
				NewStringFlag("d", "Domain (e.g. example.com)"),
				NewStringFlag("f", "Path to manifest"),
				NewIntFlag("i", "Number of instances"),
				NewStringFlag("k", "Disk quota (e.g. 256M, 1024M, 1G)"),
				NewStringFlag("m", "Memory limit (e.g. 256M, 1024M, 1G)"),
				NewStringFlag("n", "Hostname (e.g. my-subdomain)"),
				NewStringFlag("p", "Path of app directory or zip file"),
				NewStringFlag("s", "Stack to use"),
				NewStringFlag("t", "Start timeout in seconds"),
				cli.BoolFlag{Name: "no-hostname", Usage: "Map the root domain to this app"},
				cli.BoolFlag{Name: "no-manifest", Usage: "Ignore manifest file"},
				cli.BoolFlag{Name: "no-route", Usage: "Do not map a route to this app"},
				cli.BoolFlag{Name: "no-start", Usage: "Do not start an app after pushing"},
				cli.BoolFlag{Name: "random-route", Usage: "Create a random route for this app"},
			},
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("push", c)
			},
		},
		{
			Name:        "quotas",
			Description: "List available usage quotas ",
			Usage:       fmt.Sprintf("%s quotas", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("quotas", c)
			},
		},
		{
			Name:        "rename",
			Description: "Rename an app",
			Usage:       fmt.Sprintf("%s rename APP NEW_APP", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("rename", c)
			},
		},
		{
			Name:        "rename-org",
			Description: "Rename an org",
			Usage:       fmt.Sprintf("%s rename-org ORG NEW_ORG", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("rename-org", c)
			},
		},
		{
			Name:        "rename-service",
			Description: "Rename a service instance",
			Usage:       fmt.Sprintf("%s rename-service SERVICE_INSTANCE NEW_SERVICE_INSTANCE", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("rename-service", c)
			},
		},
		{
			Name:        "rename-service-broker",
			Description: "Rename a service broker",
			Usage:       fmt.Sprintf("%s rename-service-broker SERVICE_BROKER NEW_SERVICE_BROKER", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("rename-service-broker", c)
			},
		},
		{
			Name:        "rename-space",
			Description: "Rename a space",
			Usage:       fmt.Sprintf("%s rename-space SPACE NEW_SPACE", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("rename-space", c)
			},
		},
		{
			Name:        "restart",
			ShortName:   "rs",
			Description: "Restart an app",
			Usage:       fmt.Sprintf("%s restart APP", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("restart", c)
			},
		},
		{
			Name:        "routes",
			ShortName:   "r",
			Description: "List all routes in the current space",
			Usage:       fmt.Sprintf("%s routes", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("routes", c)
			},
		},
		{
			Name:        "scale",
			Description: "Change or view the instance count, disk space limit, and memory limit for an app",
			Usage:       fmt.Sprintf("%s scale APP [-i INSTANCES] [-k DISK] [-m MEMORY] [-f]", cf.Name()),
			Flags: []cli.Flag{
				NewIntFlag("i", "Number of instances"),
				NewStringFlag("k", "Disk limit (e.g. 256M, 1024M, 1G)"),
				NewStringFlag("m", "Memory limit (e.g. 256M, 1024M, 1G)"),
				cli.BoolFlag{Name: "f", Usage: "Force restart of app without prompt"},
			},
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("scale", c)
			},
		},
		{
			Name:        "service",
			Description: "Show service instance info",
			Usage:       fmt.Sprintf("%s service SERVICE_INSTANCE", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("service", c)
			},
		},
		{
			Name:        "service-auth-tokens",
			Description: "List service auth tokens",
			Usage:       fmt.Sprintf("%s service-auth-tokens", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("service-auth-tokens", c)
			},
		},
		{
			Name:        "service-brokers",
			Description: "List service brokers",
			Usage:       fmt.Sprintf("%s service-brokers", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("service-brokers", c)
			},
		},
		{
			Name:        "services",
			ShortName:   "s",
			Description: "List all services in the target space",
			Usage:       fmt.Sprintf("%s services", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("services", c)
			},
		},
		{
			Name:        "migrate-service-instances",
			Description: "Migrate service instances from one service plan to another",
			Usage: fmt.Sprintf(
				"%s migrate-service-instances v1_SERVICE v1_PROVIDER v1_PLAN v2_SERVICE v2_PLAN\n\n"+
					"WARNING: This operation is internal to Cloud Foundry; service brokers will not be contacted and"+
					" resources for service instances will not be altered. The primary use case for this operation is"+
					" to replace a service broker which implements the v1 Service Broker API with a broker which"+
					" implements the v2 API by remapping service instances from v1 plans to v2 plans.  We recommend"+
					" making the v1 plan private or shutting down the v1 broker to prevent additional instances from"+
					" being created. Once service instances have been migrated, the v1 services and plans can be"+
					" removed from Cloud Foundry.",
				cf.Name(),
			),
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "f", Usage: "Force migration without confirmation"},
			},
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("migrate-service-instances", c)
			},
		},
		{
			Name:        "set-env",
			ShortName:   "se",
			Description: "Set an env variable for an app",
			Usage:       fmt.Sprintf("%s set-env APP NAME VALUE", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("set-env", c)
			},
		},
		{
			Name:        "set-org-role",
			Description: "Assign an org role to a user",
			Usage: fmt.Sprintf("%s set-org-role USERNAME ORG ROLE\n\n", cf.Name()) +
				"ROLES:\n" +
				"   OrgManager - Invite and manage users, select and change plans, and set spending limits\n" +
				"   BillingManager - Create and manage the billing account and payment info\n" +
				"   OrgAuditor - Read-only access to org info and reports\n",
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("set-org-role", c)
			},
		},
		{
			Name:        "set-quota",
			Description: "Define the quota for an org",
			Usage: fmt.Sprintf("%s set-quota ORG QUOTA\n\n", cf.Name()) +
				"TIP:\n" +
				fmt.Sprintf("   View allowable quotas with '%s quotas'", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("set-quota", c)
			},
		},
		{
			Name:        "set-space-role",
			Description: "Assign a space role to a user",
			Usage: fmt.Sprintf("%s set-space-role USERNAME ORG SPACE ROLE\n\n", cf.Name()) +
				"ROLES:\n" +
				"   SpaceManager - Invite and manage users, and enable features for a given space\n" +
				"   SpaceDeveloper - Create and manage apps and services, and see logs and reports\n" +
				"   SpaceAuditor - View logs, reports, and settings on this space\n",
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("set-space-role", c)
			},
		},
		{
			Name:        "create-shared-domain",
			Description: "Create a domain that can be used by all orgs (admin-only)",
			Usage:       fmt.Sprintf("%s create-shared-domain DOMAIN", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("create-shared-domain", c)
			},
		},
		{
			Name:        "space",
			Description: "Show space info",
			Usage:       fmt.Sprintf("%s space SPACE", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("space", c)
			},
		},
		{
			Name:        "space-users",
			Description: "Show space users by role",
			Usage:       fmt.Sprintf("%s space-users ORG SPACE", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("space-users", c)
			},
		},
		{
			Name:        "spaces",
			Description: "List all spaces in an org",
			Usage:       fmt.Sprintf("%s spaces", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("spaces", c)
			},
		},
		{
			Name:        "stacks",
			Description: "List all stacks",
			Usage:       fmt.Sprintf("%s stacks", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("stacks", c)
			},
		},
		{
			Name:        "start",
			ShortName:   "st",
			Description: "Start an app",
			Usage:       fmt.Sprintf("%s start APP", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("start", c)
			},
		},
		{
			Name:        "stop",
			ShortName:   "sp",
			Description: "Stop an app",
			Usage:       fmt.Sprintf("%s stop APP", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("stop", c)
			},
		},
		{
			Name:        "target",
			ShortName:   "t",
			Description: "Set or view the targeted org or space",
			Usage:       fmt.Sprintf("%s target [-o ORG] [-s SPACE]", cf.Name()),
			Flags: []cli.Flag{
				NewStringFlag("o", "organization"),
				NewStringFlag("s", "space"),
			},
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("target", c)
			},
		},
		{
			Name:        "unbind-service",
			ShortName:   "us",
			Description: "Unbind a service instance from an app",
			Usage:       fmt.Sprintf("%s unbind-service APP SERVICE_INSTANCE", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("unbind-service", c)
			},
		},
		{
			Name:        "unmap-route",
			Description: "Remove a url route from an app",
			Usage:       fmt.Sprintf("%s unmap-route APP DOMAIN [-n HOSTNAME]", cf.Name()),
			Flags: []cli.Flag{
				NewStringFlag("n", "Hostname"),
			},
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("unmap-route", c)
			},
		},
		{
			Name:        "unset-env",
			Description: "Remove an env variable",
			Usage:       fmt.Sprintf("%s unset-env APP NAME", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("unset-env", c)
			},
		},
		{
			Name:        "unset-org-role",
			Description: "Remove an org role from a user",
			Usage: fmt.Sprintf("%s unset-org-role USERNAME ORG ROLE\n\n", cf.Name()) +
				"ROLES:\n" +
				"   OrgManager - Invite and manage users, select and change plans, and set spending limits\n" +
				"   BillingManager - Create and manage the billing account and payment info\n" +
				"   OrgAuditor - Read-only access to org info and reports\n",
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("unset-org-role", c)
			},
		},
		{
			Name:        "unset-space-role",
			Description: "Remove a space role from a user",
			Usage: fmt.Sprintf("%s unset-space-role USERNAME ORG SPACE ROLE\n\n", cf.Name()) +
				"ROLES:\n" +
				"   SpaceManager - Invite and manage users, and enable features for a given space\n" +
				"   SpaceDeveloper - Create and manage apps and services, and see logs and reports\n" +
				"   SpaceAuditor - View logs, reports, and settings on this space\n",
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("unset-space-role", c)
			},
		},
		{
			Name:        "update-buildpack",
			Description: "Update a buildpack",
			Usage:       fmt.Sprintf("%s update-buildpack BUILDPACK [-p PATH] [-i POSITION] [--enable|--disable] [--lock|--unlock]", cf.Name()),
			Flags: []cli.Flag{
				NewIntFlag("i", "Buildpack position among other buildpacks"),
				NewStringFlag("p", "Path to directory or zip file"),
				cli.BoolFlag{Name: "enable", Usage: "Enable the buildpack"},
				cli.BoolFlag{Name: "disable", Usage: "Disable the buildpack"},
				cli.BoolFlag{Name: "lock", Usage: "Lock the buildpack"},
				cli.BoolFlag{Name: "unlock", Usage: "Unlock the buildpack"},
			},
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("update-buildpack", c)
			},
		},
		{
			Name:        "update-service-broker",
			Description: "Update a service broker",
			Usage:       fmt.Sprintf("%s update-service-broker SERVICE_BROKER USERNAME PASSWORD URL", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("update-service-broker", c)
			},
		},
		{
			Name:        "update-service-auth-token",
			Description: "Update a service auth token",
			Usage:       fmt.Sprintf("%s update-service-auth-token LABEL PROVIDER TOKEN", cf.Name()),
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("update-service-auth-token", c)
			},
		},
		{
			Name:        "update-user-provided-service",
			ShortName:   "uups",
			Description: "Update user-provided service name value pairs",
			Usage: fmt.Sprintf("%s update-user-provided-service SERVICE_INSTANCE [-p PARAMETERS] [-l SYSLOG-DRAIN-URL]'\n\n", cf.Name()) +
				"EXAMPLE:\n" +
				fmt.Sprintf("   %s update-user-provided-service oracle-db-mine -p '{\"username\":\"admin\",\"password\":\"pa55woRD\"}'\n", cf.Name()) +
				fmt.Sprintf("   %s update-user-provided-service my-drain-service -l syslog://example.com\n", cf.Name()),
			Flags: []cli.Flag{
				NewStringFlag("p", "Parameters"),
				NewStringFlag("l", "Syslog Drain Url"),
			},
			Action: func(c *cli.Context) {
				cmdRunner.RunCmdByName("update-user-provided-service", c)
			},
		},
	}
	return
}
