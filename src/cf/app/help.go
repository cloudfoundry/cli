package app

import (
	"github.com/codegangsta/cli"
	"text/template"
	"os"
	"text/tabwriter"
	"cf/terminal"
)

var appHelpTemplate = `{{.Title "NAME:"}}
   {{.Name}} - {{.Usage}}

{{.Title "USAGE:"}}
   [environment variables] {{.Name}} [global options] command [arguments...] [command options]

{{.Title "VERSION:"}}
   {{.Version}}

{{.Title "COMMANDS:"}}
   {{range .Commands}}
{{.SubTitle .Name}}
{{range .CommandSubGroups}}
{{range .}}   {{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Description}}
{{end}}{{end}}{{end}}
{{.Title "GLOBAL OPTIONS:"}}
   {{range .Flags}}{{.}}
   {{end}}
{{.Title "ENVIRONMENT VARIABLES:"}}
   CF_TRACE=true - will output HTTP requests and responses during command
   HTTP_PROXY=http://proxy.example.com:8080 - set to your proxy
`

type appPresenter struct {
	cli.App
	Commands []groupedCommands
}

func (p appPresenter) Title(name string) string {
	return terminal.HeaderColor(name)
}

type groupedCommands struct {
	Name             string
	CommandSubGroups [][]*cli.Command
}

func (c groupedCommands) SubTitle(name string) string {
	return terminal.TableContentHeaderColor(name)
}

func newAppPresenter(app *cli.App) (presenter appPresenter) {
	presenter.Name = app.Name
	presenter.Usage = app.Usage
	presenter.Version = app.Version
	presenter.Name = app.Name
	presenter.Flags = app.Flags

	presenter.Commands = []groupedCommands{
		{
			Name: "GETTING STARTED",
			CommandSubGroups: [][]*cli.Command{
				{
					app.Command("login"),
					app.Command("logout"),
					app.Command("passwd"),
					app.Command("target"),
				}, {
					app.Command("api"),
					app.Command("auth"),
				},
			},
		}, {
			Name: "APPS",
			CommandSubGroups: [][]*cli.Command{
				{
					app.Command("apps"),
					app.Command("app"),
				}, {
					app.Command("push"),
					app.Command("scale"),
					app.Command("delete"),
					app.Command("rename"),
				}, {
					app.Command("start"),
					app.Command("stop"),
					app.Command("restart"),
				}, {
					app.Command("events"),
					app.Command("files"),
					app.Command("logs"),
				}, {
					app.Command("env"),
					app.Command("set-env"),
					app.Command("unset-env"),
				}, {
					app.Command("buildpacks"),
					app.Command("stacks"),
				},
			},
		}, {
			Name: "SERVICES",
			CommandSubGroups: [][]*cli.Command{
				{
					app.Command("marketplace"),
					app.Command("services"),
					app.Command("service"),
				}, {
					app.Command("create-service"),
					app.Command("delete-service"),
					app.Command("rename-service"),
				}, {
					app.Command("bind-service"),
					app.Command("unbind-service"),
				}, {
					app.Command("create-user-provided-service"),
					app.Command("update-user-provided-service"),
				},
			},
		}, {
			Name: "ORGS",
			CommandSubGroups: [][]*cli.Command{
				{
					app.Command("orgs"),
					app.Command("org"),
				}, {
					app.Command("create-org"),
					app.Command("delete-org"),
					app.Command("rename-org"),
				},
			},
		}, {
			Name: "SPACES",
			CommandSubGroups: [][]*cli.Command{
				{
					app.Command("spaces"),
					app.Command("space"),
				}, {
					app.Command("create-space"),
					app.Command("delete-space"),
					app.Command("rename-space"),
				},
			},
		}, {
			Name: "DOMAINS",
			CommandSubGroups: [][]*cli.Command{
				{
					app.Command("domains"),
					app.Command("share-domain"),
					app.Command("reserve-domain"),
					app.Command("map-domain"),
					app.Command("unmap-domain"),
					app.Command("delete-domain"),
				},
			},
		}, {
			Name: "ROUTES",
			CommandSubGroups: [][]*cli.Command{
				{
					app.Command("routes"),
					app.Command("reserve-route"),
					app.Command("map-route"),
					app.Command("unmap-route"),
					app.Command("delete-route"),
				},
			},
		}, {
			Name: "USER ADMIN",
			CommandSubGroups: [][]*cli.Command{
				{
					app.Command("create-user"),
					app.Command("delete-user"),
				}, {
					app.Command("org-users"),
					app.Command("set-org-role"),
					app.Command("unset-org-role"),
				}, {
					app.Command("space-users"),
					app.Command("set-space-role"),
					app.Command("unset-space-role"),
				},
			},
		}, {
			Name: "ORG ADMIN",
			CommandSubGroups: [][]*cli.Command{
				{
					app.Command("quotas"),
					app.Command("set-quota"),
				},
			},
		}, {
			Name: "SERVICE ADMIN",
			CommandSubGroups: [][]*cli.Command{
				{
					app.Command("service-auth-tokens"),
					app.Command("create-service-auth-token"),
					app.Command("update-service-auth-token"),
					app.Command("delete-service-auth-token"),
				}, {
					app.Command("service-brokers"),
					app.Command("create-service-broker"),
					app.Command("update-service-broker"),
					app.Command("delete-service-broker"),
					app.Command("rename-service-broker"),
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
