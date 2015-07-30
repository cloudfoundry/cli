package app

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"text/tabwriter"
	"text/template"
	"unicode/utf8"

	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/plugin_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
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
	maxNameLen := command_registry.Commands.MaxCommandNameLength()

	presentCommand := func(commandName string) (presenter cmdPresenter) {
		cmd := app.Command(commandName)
		presenter.Name = presentCmdName(*cmd)
		padding := strings.Repeat(" ", maxNameLen-utf8.RuneCountInString(presenter.Name))
		presenter.Name = presenter.Name + padding
		presenter.Description = cmd.Description
		return
	}

	presentNonCodegangstaCommand := func(commandName string) (presenter cmdPresenter) {
		cmd := command_registry.Commands.FindCommand(commandName)
		presenter.Name = cmd.MetaData().Name
		padding := strings.Repeat(" ", maxNameLen-utf8.RuneCountInString(presenter.Name))
		presenter.Name = presenter.Name + padding
		presenter.Description = cmd.MetaData().Description
		return
	}

	presentPluginCommands := func() []cmdPresenter {
		pluginConfig := plugin_config.NewPluginConfig(func(err error) {
			//fail silently when running help?
		})

		plugins := pluginConfig.Plugins()
		var presenters []cmdPresenter
		var pluginPresenter cmdPresenter

		for _, pluginMetadata := range plugins {
			for _, cmd := range pluginMetadata.Commands {

				if cmd.Alias == "" {
					pluginPresenter.Name = cmd.Name
				} else {
					pluginPresenter.Name = cmd.Name + ", " + cmd.Alias
				}

				padding := strings.Repeat(" ", maxNameLen-utf8.RuneCountInString(pluginPresenter.Name))
				pluginPresenter.Name = pluginPresenter.Name + padding
				pluginPresenter.Description = cmd.HelpText
				presenters = append(presenters, pluginPresenter)
			}
		}

		return presenters
	}
	presenter.Name = app.Name
	presenter.Flags = app.Flags
	presenter.Usage = app.Usage
	presenter.Version = app.Version
	presenter.Compiled = app.Compiled
	presenter.Commands = []groupedCommands{
		{
			Name: T("GETTING STARTED"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentNonCodegangstaCommand("login"),
					presentNonCodegangstaCommand("logout"),
					presentNonCodegangstaCommand("passwd"),
					presentNonCodegangstaCommand("target"),
				}, {
					presentNonCodegangstaCommand("api"),
					presentNonCodegangstaCommand("auth"),
				},
			},
		}, {
			Name: T("APPS"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentNonCodegangstaCommand("apps"),
					presentNonCodegangstaCommand("app"),
				}, {
					presentCommand("push"),                //needs start/restart ...
					presentNonCodegangstaCommand("scale"), //needs stop/restart
					presentNonCodegangstaCommand("delete"),
					presentNonCodegangstaCommand("rename"),
				}, {
					presentNonCodegangstaCommand("start"), //needs app
					presentNonCodegangstaCommand("stop"),
					presentNonCodegangstaCommand("restart"), //needs start
					presentNonCodegangstaCommand("restage"),
					presentCommand("restart-app-instance"),
				}, {
					presentNonCodegangstaCommand("events"),
					presentNonCodegangstaCommand("files"),
					presentNonCodegangstaCommand("logs"),
				}, {
					presentNonCodegangstaCommand("env"),
					presentNonCodegangstaCommand("set-env"),
					presentNonCodegangstaCommand("unset-env"),
				}, {
					presentNonCodegangstaCommand("stacks"),
					presentNonCodegangstaCommand("stack"),
				}, {
					presentNonCodegangstaCommand("copy-source"),
				}, {
					presentNonCodegangstaCommand("create-app-manifest"),
				},
			},
		}, {
			Name: T("SERVICES"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentNonCodegangstaCommand("marketplace"),
					presentNonCodegangstaCommand("services"),
					presentNonCodegangstaCommand("service"),
				}, {
					presentNonCodegangstaCommand("create-service"),
					presentCommand("update-service"),
					presentNonCodegangstaCommand("delete-service"),
					presentNonCodegangstaCommand("rename-service"),
				}, {
					presentNonCodegangstaCommand("create-service-key"),
					presentNonCodegangstaCommand("service-keys"),
					presentCommand("service-key"),
					presentCommand("delete-service-key"),
				}, {
					presentNonCodegangstaCommand("bind-service"),
					presentCommand("unbind-service"),
				}, {
					presentNonCodegangstaCommand("create-user-provided-service"),
					presentCommand("update-user-provided-service"),
				},
			},
		}, {
			Name: T("ORGS"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentNonCodegangstaCommand("orgs"),
					presentNonCodegangstaCommand("org"),
				}, {
					presentNonCodegangstaCommand("create-org"),
					presentNonCodegangstaCommand("delete-org"),
					presentNonCodegangstaCommand("rename-org"),
				},
			},
		}, {
			Name: T("SPACES"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentNonCodegangstaCommand("spaces"),
					presentNonCodegangstaCommand("space"),
				}, {
					presentNonCodegangstaCommand("create-space"),
					presentNonCodegangstaCommand("delete-space"),
					presentNonCodegangstaCommand("rename-space"),
				},
			},
		}, {
			Name: T("DOMAINS"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentNonCodegangstaCommand("domains"),
					presentNonCodegangstaCommand("create-domain"),
					presentNonCodegangstaCommand("delete-domain"),
					presentNonCodegangstaCommand("create-shared-domain"),
					presentNonCodegangstaCommand("delete-shared-domain"),
				},
			},
		}, {
			Name: T("ROUTES"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentNonCodegangstaCommand("routes"),
					presentNonCodegangstaCommand("create-route"),
					presentNonCodegangstaCommand("check-route"),
					presentNonCodegangstaCommand("map-route"),
					presentNonCodegangstaCommand("unmap-route"),
					presentNonCodegangstaCommand("delete-route"),
					presentNonCodegangstaCommand("delete-orphaned-routes"),
				},
			},
		}, {
			Name: T("BUILDPACKS"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentNonCodegangstaCommand("buildpacks"),
					presentNonCodegangstaCommand("create-buildpack"),
					presentNonCodegangstaCommand("update-buildpack"),
					presentNonCodegangstaCommand("rename-buildpack"),
					presentNonCodegangstaCommand("delete-buildpack"),
				},
			},
		}, {
			Name: T("USER ADMIN"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentCommand("create-user"),
					presentNonCodegangstaCommand("delete-user"),
				}, {
					presentNonCodegangstaCommand("org-users"),
					presentNonCodegangstaCommand("set-org-role"),
					presentNonCodegangstaCommand("unset-org-role"),
				}, {
					presentNonCodegangstaCommand("space-users"),
					presentNonCodegangstaCommand("set-space-role"),
					presentNonCodegangstaCommand("unset-space-role"),
				},
			},
		}, {
			Name: T("ORG ADMIN"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentNonCodegangstaCommand("quotas"),
					presentNonCodegangstaCommand("quota"),
					presentNonCodegangstaCommand("set-quota"),
				}, {
					presentNonCodegangstaCommand("create-quota"),
					presentNonCodegangstaCommand("delete-quota"),
					presentNonCodegangstaCommand("update-quota"),
				},
				{
					presentNonCodegangstaCommand("share-private-domain"),
					presentNonCodegangstaCommand("unshare-private-domain"),
				},
			},
		}, {
			Name: T("SPACE ADMIN"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentNonCodegangstaCommand("space-quotas"),
					presentNonCodegangstaCommand("space-quota"),
					presentNonCodegangstaCommand("create-space-quota"),
					presentNonCodegangstaCommand("update-space-quota"),
					presentNonCodegangstaCommand("delete-space-quota"),
					presentNonCodegangstaCommand("set-space-quota"),
					presentNonCodegangstaCommand("unset-space-quota"),
				},
			},
		}, {
			Name: T("SERVICE ADMIN"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentNonCodegangstaCommand("service-auth-tokens"),
					presentNonCodegangstaCommand("create-service-auth-token"),
					presentNonCodegangstaCommand("update-service-auth-token"),
					presentNonCodegangstaCommand("delete-service-auth-token"),
				}, {
					presentNonCodegangstaCommand("service-brokers"),
					presentNonCodegangstaCommand("create-service-broker"),
					presentNonCodegangstaCommand("update-service-broker"),
					presentNonCodegangstaCommand("delete-service-broker"),
					presentNonCodegangstaCommand("rename-service-broker"),
				}, {
					presentNonCodegangstaCommand("migrate-service-instances"),
					presentNonCodegangstaCommand("purge-service-offering"),
				}, {
					presentNonCodegangstaCommand("service-access"),
					presentCommand("enable-service-access"),
					presentCommand("disable-service-access"),
				},
			},
		}, {
			Name: T("SECURITY GROUP"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentNonCodegangstaCommand("security-group"),
					presentNonCodegangstaCommand("security-groups"),
					presentNonCodegangstaCommand("create-security-group"),
					presentNonCodegangstaCommand("update-security-group"),
					presentNonCodegangstaCommand("delete-security-group"),
					presentNonCodegangstaCommand("bind-security-group"),
					presentNonCodegangstaCommand("unbind-security-group"),
				}, {
					presentNonCodegangstaCommand("bind-staging-security-group"),
					presentNonCodegangstaCommand("staging-security-groups"),
					presentNonCodegangstaCommand("unbind-staging-security-group"),
				}, {
					presentNonCodegangstaCommand("bind-running-security-group"),
					presentNonCodegangstaCommand("running-security-groups"),
					presentNonCodegangstaCommand("unbind-running-security-group"),
				},
			},
		}, {
			Name: T("ENVIRONMENT VARIABLE GROUPS"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentNonCodegangstaCommand("running-environment-variable-group"),
					presentNonCodegangstaCommand("staging-environment-variable-group"),
					presentNonCodegangstaCommand("set-staging-environment-variable-group"),
					presentNonCodegangstaCommand("set-running-environment-variable-group"),
				},
			},
		},
		{
			Name: T("FEATURE FLAGS"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentNonCodegangstaCommand("feature-flags"),
					presentNonCodegangstaCommand("feature-flag"),
					presentNonCodegangstaCommand("enable-feature-flag"),
					presentNonCodegangstaCommand("disable-feature-flag"),
				},
			},
		}, {
			Name: T("ADVANCED"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentNonCodegangstaCommand("curl"),
					presentNonCodegangstaCommand("config"),
					presentNonCodegangstaCommand("oauth-token"),
				},
			},
		}, {
			Name: T("ADD/REMOVE PLUGIN REPOSITORY"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentNonCodegangstaCommand("add-plugin-repo"),
					presentNonCodegangstaCommand("remove-plugin-repo"),
					presentNonCodegangstaCommand("list-plugin-repos"),
					presentNonCodegangstaCommand("repo-plugins"),
				},
			},
		}, {
			Name: T("ADD/REMOVE PLUGIN"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentNonCodegangstaCommand("plugins"),
					presentCommand("install-plugin"),
					presentNonCodegangstaCommand("uninstall-plugin"),
				},
			},
		}, {
			Name: T("INSTALLED PLUGIN COMMANDS"),
			CommandSubGroups: [][]cmdPresenter{
				presentPluginCommands(),
			},
		},
	}

	return
}

func ShowHelp(helpTemplate string, thingToPrint interface{}) {
	translatedTemplatedHelp := T(strings.Replace(helpTemplate, "{{", "[[", -1))
	translatedTemplatedHelp = strings.Replace(translatedTemplatedHelp, "[[", "{{", -1)

	switch thing := thingToPrint.(type) {
	case *cli.App:
		showAppHelp(translatedTemplatedHelp, thing)
	case cli.Command:
		showCommandHelp(translatedTemplatedHelp, thing)
	default:
		panic(fmt.Sprintf("Help printer has received something that is neither app nor command! The beast (%s) looks like this: %s", reflect.TypeOf(thing), thing))
	}
}

var CodeGangstaHelpPrinter = cli.HelpPrinter

func showCommandHelp(helpTemplate string, commandToPrint cli.Command) {
	CodeGangstaHelpPrinter(helpTemplate, commandToPrint)
}

func showAppHelp(helpTemplate string, appToPrint *cli.App) {
	presenter := newAppPresenter(appToPrint)

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	t := template.Must(template.New("help").Parse(helpTemplate))
	t.Execute(w, presenter)
	w.Flush()
}
