package app

import (
	"os"
	"strings"
	"text/tabwriter"
	"text/template"
	"unicode/utf8"

	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

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

func newAppPresenter(app *cli.App) (presenter appPresenter) {
	maxNameLen := 0
	for _, cmd := range app.Commands {
		name := presentCmdName(cmd)
		if utf8.RuneCountInString(name) > maxNameLen {
			maxNameLen = len(name)
		}
	}

	presentCommand := func(commandName string) (presenter cmdPresenter) {
		cmd := app.Command(commandName)
		presenter.Name = presentCmdName(*cmd)
		padding := strings.Repeat(" ", maxNameLen-utf8.RuneCountInString(presenter.Name))
		presenter.Name = presenter.Name + padding
		presenter.Description = cmd.Description
		return
	}

	presenter.Name = app.Name
	presenter.Usage = app.Usage
	presenter.Version = app.Version
	presenter.Name = app.Name
	presenter.Flags = app.Flags
	presenter.Compiled = app.Compiled

	presenter.Commands = []groupedCommands{
		{
			Name: T("GETTING STARTED"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentCommand("login"),
					presentCommand("logout"),
					presentCommand("passwd"),
					presentCommand("target"),
				}, {
					presentCommand("api"),
					presentCommand("auth"),
				},
			},
		}, {
			Name: T("APPS"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentCommand("apps"),
					presentCommand("app"),
				}, {
					presentCommand("push"),
					presentCommand("scale"),
					presentCommand("delete"),
					presentCommand("rename"),
				}, {
					presentCommand("start"),
					presentCommand("stop"),
					presentCommand("restart"),
					presentCommand("restage"),
				}, {
					presentCommand("events"),
					presentCommand("files"),
					presentCommand("logs"),
				}, {
					presentCommand("env"),
					presentCommand("set-env"),
					presentCommand("unset-env"),
				}, {
					presentCommand("stacks"),
				},
			},
		}, {
			Name: T("SERVICES"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentCommand("marketplace"),
					presentCommand("services"),
					presentCommand("service"),
				}, {
					presentCommand("create-service"),
					presentCommand("delete-service"),
					presentCommand("rename-service"),
				}, {
					presentCommand("bind-service"),
					presentCommand("unbind-service"),
				}, {
					presentCommand("create-user-provided-service"),
					presentCommand("update-user-provided-service"),
				},
			},
		}, {
			Name: T("ORGS"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentCommand("orgs"),
					presentCommand("org"),
				}, {
					presentCommand("create-org"),
					presentCommand("delete-org"),
					presentCommand("rename-org"),
				},
			},
		}, {
			Name: T("SPACES"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentCommand("spaces"),
					presentCommand("space"),
				}, {
					presentCommand("create-space"),
					presentCommand("delete-space"),
					presentCommand("rename-space"),
				},
			},
		}, {
			Name: T("DOMAINS"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentCommand("domains"),
					presentCommand("create-domain"),
					presentCommand("delete-domain"),
					presentCommand("create-shared-domain"),
					presentCommand("delete-shared-domain"),
				},
			},
		}, {
			Name: T("ROUTES"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentCommand("routes"),
					presentCommand("create-route"),
					presentCommand("map-route"),
					presentCommand("unmap-route"),
					presentCommand("delete-route"),
					presentCommand("delete-orphaned-routes"),
				},
			},
		}, {
			Name: T("BUILDPACKS"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentCommand("buildpacks"),
					presentCommand("create-buildpack"),
					presentCommand("update-buildpack"),
					presentCommand("rename-buildpack"),
					presentCommand("delete-buildpack"),
				},
			},
		}, {
			Name: T("USER ADMIN"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentCommand("create-user"),
					presentCommand("delete-user"),
				}, {
					presentCommand("org-users"),
					presentCommand("set-org-role"),
					presentCommand("unset-org-role"),
				}, {
					presentCommand("space-users"),
					presentCommand("set-space-role"),
					presentCommand("unset-space-role"),
				},
			},
		}, {
			Name: T("ORG ADMIN"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentCommand("quotas"),
					presentCommand("quota"),
					presentCommand("set-quota"),
				}, {
					presentCommand("create-quota"),
					presentCommand("delete-quota"),
					presentCommand("update-quota"),
				},
			},
		}, {
			Name: T("SERVICE ADMIN"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentCommand("service-auth-tokens"),
					presentCommand("create-service-auth-token"),
					presentCommand("update-service-auth-token"),
					presentCommand("delete-service-auth-token"),
				}, {
					presentCommand("service-brokers"),
					presentCommand("create-service-broker"),
					presentCommand("update-service-broker"),
					presentCommand("delete-service-broker"),
					presentCommand("rename-service-broker"),
				}, {
					presentCommand("migrate-service-instances"),
					presentCommand("purge-service-offering"),
				},
			},
		}, {
			Name: T("APPLICATION SECURITY GROUP"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentCommand("application-security-group"),
					presentCommand("create-application-security-group"),
					presentCommand("delete-application-security-group"),
					presentCommand("add-default-staging-application-security-group"),
				},
			},
		}, {
			Name: T("ADVANCED"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentCommand("curl"),
					presentCommand("config"),
				},
			},
		},
	}
	return
}

func ShowAppHelp(helpTemplate string, appToPrint interface{}) {
	translatedTemplatedHelp := T(strings.Replace(helpTemplate, "{{", "[[", -1))
	translatedTemplatedHelp = strings.Replace(helpTemplate, "[[", "{{", -1)
	showAppHelp(translatedTemplatedHelp, appToPrint)
}

func showAppHelp(helpTemplate string, appToPrint interface{}) {
	app := appToPrint.(*cli.App)
	presenter := newAppPresenter(app)

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	t := template.Must(template.New("help").Parse(helpTemplate))
	t.Execute(w, presenter)
	w.Flush()
}
