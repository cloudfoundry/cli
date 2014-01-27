package app

import (
	"cf/terminal"
	"github.com/codegangsta/cli"
	"os"
	"strings"
	"text/tabwriter"
	"text/template"
)

var appHelpTemplate = `{{.Title "NAME:"}}
   {{.Name}} - {{.Usage}}

{{.Title "USAGE:"}}
   [environment variables] {{.Name}} [global options] command [arguments...] [command options]

{{.Title "VERSION:"}}
   {{.Version}}
   {{range .Commands}}
{{.SubTitle .Name}}{{range .CommandSubGroups}}
{{range .}}   {{.Name}} {{.Description}}
{{end}}{{end}}{{end}}
{{.Title "GLOBAL OPTIONS:"}}
   {{range .Flags}}{{.}}
   {{end}}
{{.Title "ENVIRONMENT VARIABLES:"}}
   CF_HOME=path/to/config/ override default config directory
   CF_STAGING_TIMEOUT=15 max wait time for buildpack staging, in minutes
   CF_STARTUP_TIMEOUT=5 max wait time for app instance startup, in minutes
   CF_COLOR=false - will not colorize output
   CF_TRACE=true - print API request diagnostics to stdout
   CF_TRACE=path/to/trace.log - append API request diagnostics to a log file
   HTTP_PROXY=http://proxy.example.com:8080 - enable HTTP proxying for API requests
`

type groupedCommands struct {
	Name             string
	CommandSubGroups [][]cmdPresenter
}

func (c groupedCommands) SubTitle(name string) string {
	return terminal.HeaderColor(name + ":")
}

type cmdPresenter struct {
	Name        string
	Description string
}

func newCmdPresenter(app *cli.App, maxNameLen int, cmdName string) (presenter cmdPresenter) {
	cmd := app.Command(cmdName)

	presenter.Name = presentCmdName(*cmd)
	padding := strings.Repeat(" ", maxNameLen-len(presenter.Name))
	presenter.Name = presenter.Name + padding

	presenter.Description = cmd.Description

	return
}

func presentCmdName(cmd cli.Command) (name string) {
	name = cmd.Name
	if cmd.ShortName != "" {
		name = name + ", " + cmd.ShortName
	}
	return
}

type appPresenter struct {
	cli.App
	Commands []groupedCommands
}

func (p appPresenter) Title(name string) string {
	return terminal.HeaderColor(name)
}

func getMaxCmdNameLength(app *cli.App) (length int) {
	for _, cmd := range app.Commands {
		name := presentCmdName(cmd)
		if len(name) > length {
			length = len(name)
		}
	}
	return
}

func newAppPresenter(app *cli.App) (presenter appPresenter) {
	maxNameLen := getMaxCmdNameLength(app)

	presenter.Name = app.Name
	presenter.Usage = app.Usage
	presenter.Version = app.Version
	presenter.Name = app.Name
	presenter.Flags = app.Flags

	presenter.Commands = []groupedCommands{
		{
			Name: "GETTING STARTED",
			CommandSubGroups: [][]cmdPresenter{
				{
					newCmdPresenter(app, maxNameLen, "login"),
					newCmdPresenter(app, maxNameLen, "logout"),
					newCmdPresenter(app, maxNameLen, "passwd"),
					newCmdPresenter(app, maxNameLen, "target"),
				}, {
					newCmdPresenter(app, maxNameLen, "api"),
					newCmdPresenter(app, maxNameLen, "auth"),
				},
			},
		}, {
			Name: "APPS",
			CommandSubGroups: [][]cmdPresenter{
				{
					newCmdPresenter(app, maxNameLen, "apps"),
					newCmdPresenter(app, maxNameLen, "app"),
				}, {
					newCmdPresenter(app, maxNameLen, "push"),
					newCmdPresenter(app, maxNameLen, "scale"),
					newCmdPresenter(app, maxNameLen, "delete"),
					newCmdPresenter(app, maxNameLen, "rename"),
				}, {
					newCmdPresenter(app, maxNameLen, "start"),
					newCmdPresenter(app, maxNameLen, "stop"),
					newCmdPresenter(app, maxNameLen, "restart"),
				}, {
					newCmdPresenter(app, maxNameLen, "events"),
					newCmdPresenter(app, maxNameLen, "files"),
					newCmdPresenter(app, maxNameLen, "logs"),
				}, {
					newCmdPresenter(app, maxNameLen, "env"),
					newCmdPresenter(app, maxNameLen, "set-env"),
					newCmdPresenter(app, maxNameLen, "unset-env"),
				}, {
					newCmdPresenter(app, maxNameLen, "stacks"),
				},
			},
		}, {
			Name: "SERVICES",
			CommandSubGroups: [][]cmdPresenter{
				{
					newCmdPresenter(app, maxNameLen, "marketplace"),
					newCmdPresenter(app, maxNameLen, "services"),
					newCmdPresenter(app, maxNameLen, "service"),
				}, {
					newCmdPresenter(app, maxNameLen, "create-service"),
					newCmdPresenter(app, maxNameLen, "delete-service"),
					newCmdPresenter(app, maxNameLen, "rename-service"),
				}, {
					newCmdPresenter(app, maxNameLen, "bind-service"),
					newCmdPresenter(app, maxNameLen, "unbind-service"),
				}, {
					newCmdPresenter(app, maxNameLen, "create-user-provided-service"),
					newCmdPresenter(app, maxNameLen, "update-user-provided-service"),
				},
			},
		}, {
			Name: "ORGS",
			CommandSubGroups: [][]cmdPresenter{
				{
					newCmdPresenter(app, maxNameLen, "orgs"),
					newCmdPresenter(app, maxNameLen, "org"),
				}, {
					newCmdPresenter(app, maxNameLen, "create-org"),
					newCmdPresenter(app, maxNameLen, "delete-org"),
					newCmdPresenter(app, maxNameLen, "rename-org"),
				},
			},
		}, {
			Name: "SPACES",
			CommandSubGroups: [][]cmdPresenter{
				{
					newCmdPresenter(app, maxNameLen, "spaces"),
					newCmdPresenter(app, maxNameLen, "space"),
				}, {
					newCmdPresenter(app, maxNameLen, "create-space"),
					newCmdPresenter(app, maxNameLen, "delete-space"),
					newCmdPresenter(app, maxNameLen, "rename-space"),
				},
			},
		}, {
			Name: "DOMAINS",
			CommandSubGroups: [][]cmdPresenter{
				{
					newCmdPresenter(app, maxNameLen, "domains"),
					newCmdPresenter(app, maxNameLen, "create-domain"),
					newCmdPresenter(app, maxNameLen, "share-domain"),
					newCmdPresenter(app, maxNameLen, "delete-domain"),
				},
			},
		}, {
			Name: "ROUTES",
			CommandSubGroups: [][]cmdPresenter{
				{
					newCmdPresenter(app, maxNameLen, "routes"),
					newCmdPresenter(app, maxNameLen, "create-route"),
					newCmdPresenter(app, maxNameLen, "map-route"),
					newCmdPresenter(app, maxNameLen, "unmap-route"),
					newCmdPresenter(app, maxNameLen, "delete-route"),
				},
			},
		}, {
			Name: "BUILDPACKS",
			CommandSubGroups: [][]cmdPresenter{
				{
					newCmdPresenter(app, maxNameLen, "buildpacks"),
					newCmdPresenter(app, maxNameLen, "create-buildpack"),
					newCmdPresenter(app, maxNameLen, "update-buildpack"),
					newCmdPresenter(app, maxNameLen, "delete-buildpack"),
				},
			},
		}, {
			Name: "USER ADMIN",
			CommandSubGroups: [][]cmdPresenter{
				{
					newCmdPresenter(app, maxNameLen, "create-user"),
					newCmdPresenter(app, maxNameLen, "delete-user"),
				}, {
					newCmdPresenter(app, maxNameLen, "org-users"),
					newCmdPresenter(app, maxNameLen, "set-org-role"),
					newCmdPresenter(app, maxNameLen, "unset-org-role"),
				}, {
					newCmdPresenter(app, maxNameLen, "space-users"),
					newCmdPresenter(app, maxNameLen, "set-space-role"),
					newCmdPresenter(app, maxNameLen, "unset-space-role"),
				},
			},
		}, {
			Name: "ORG ADMIN",
			CommandSubGroups: [][]cmdPresenter{
				{
					newCmdPresenter(app, maxNameLen, "quotas"),
					newCmdPresenter(app, maxNameLen, "set-quota"),
				},
			},
		}, {
			Name: "SERVICE ADMIN",
			CommandSubGroups: [][]cmdPresenter{
				{
					newCmdPresenter(app, maxNameLen, "service-auth-tokens"),
					newCmdPresenter(app, maxNameLen, "create-service-auth-token"),
					newCmdPresenter(app, maxNameLen, "update-service-auth-token"),
					newCmdPresenter(app, maxNameLen, "delete-service-auth-token"),
				}, {
					newCmdPresenter(app, maxNameLen, "service-brokers"),
					newCmdPresenter(app, maxNameLen, "create-service-broker"),
					newCmdPresenter(app, maxNameLen, "update-service-broker"),
					newCmdPresenter(app, maxNameLen, "delete-service-broker"),
					newCmdPresenter(app, maxNameLen, "rename-service-broker"),
				},
			},
		}, {
			Name: "ADVANCED",
			CommandSubGroups: [][]cmdPresenter{
				{
					newCmdPresenter(app, maxNameLen, "curl"),
				},
			},
		},
	}
	return
}

func showAppHelp(app *cli.App) {
	presenter := newAppPresenter(app)

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	t := template.Must(template.New("help").Parse(appHelpTemplate))
	t.Execute(w, presenter)
	w.Flush()
}
