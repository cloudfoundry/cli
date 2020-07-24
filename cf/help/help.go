package help

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"text/template"
	"unicode/utf8"

	"path/filepath"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration"
	"code.cloudfoundry.org/cli/cf/configuration/confighelpers"
	"code.cloudfoundry.org/cli/cf/configuration/pluginconfig"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/version"
)

type appPresenter struct {
	Name     string
	Usage    string
	Version  string
	Compiled string
	Commands []groupedCommands
}

type groupedCommands struct {
	Name             string
	CommandSubGroups [][]cmdPresenter
}

type cmdPresenter struct {
	Name        string
	Description string
}

func ShowHelp(writer io.Writer, helpTemplate string) {
	translatedTemplatedHelp := T(strings.Replace(helpTemplate, "{{", "[[", -1))
	translatedTemplatedHelp = strings.Replace(translatedTemplatedHelp, "[[", "{{", -1)

	showAppHelp(writer, translatedTemplatedHelp)
}

func showAppHelp(writer io.Writer, helpTemplate string) {
	presenter := newAppPresenter()

	w := tabwriter.NewWriter(writer, 0, 8, 1, '\t', 0)
	t := template.Must(template.New("help").Parse(helpTemplate))
	err := t.Execute(w, presenter)
	if err != nil {
		fmt.Println("error", err)
	}
	_ = w.Flush()
}

func newAppPresenter() appPresenter {
	var presenter appPresenter

	pluginPath := filepath.Join(confighelpers.PluginRepoDir(), ".cf", "plugins")

	pluginConfig := pluginconfig.NewPluginConfig(
		func(err error) {
			//fail silently when running help
		},
		configuration.NewDiskPersistor(filepath.Join(pluginPath, "config.json")),
		pluginPath,
	)

	plugins := pluginConfig.Plugins()

	maxNameLen := commandregistry.Commands.MaxCommandNameLength()
	maxNameLen = maxPluginCommandNameLength(plugins, maxNameLen)

	presentCommand := func(commandName string) (presenter cmdPresenter) {
		cmd := commandregistry.Commands.FindCommand(commandName)
		presenter.Name = cmd.MetaData().Name
		padding := strings.Repeat(" ", maxNameLen-utf8.RuneCountInString(presenter.Name))
		presenter.Name = presenter.Name + padding
		presenter.Description = cmd.MetaData().Description
		return
	}

	presentPluginCommands := func() []cmdPresenter {
		var presenters []cmdPresenter
		var pluginPresenter cmdPresenter

		for _, pluginMetadata := range plugins {
			for _, cmd := range pluginMetadata.Commands {
				pluginPresenter.Name = cmd.Name

				padding := strings.Repeat(" ", maxNameLen-utf8.RuneCountInString(pluginPresenter.Name))
				pluginPresenter.Name = pluginPresenter.Name + padding
				pluginPresenter.Description = cmd.HelpText
				presenters = append(presenters, pluginPresenter)
			}
		}

		return presenters
	}

	presenter.Name = os.Args[0]
	presenter.Usage = T("A command line tool to interact with Cloud Foundry")
	presenter.Version = version.VersionString()
	presenter.Commands = []groupedCommands{
		{
			Name: T("GETTING STARTED"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentCommand("help"),
					presentCommand("version"),
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
					presentCommand("restart-app-instance"),
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
					presentCommand("stack"),
				}, {
					presentCommand("copy-source"),
				}, {
					presentCommand("create-app-manifest"),
				}, {
					presentCommand("get-health-check"),
					presentCommand("set-health-check"),
					presentCommand("enable-ssh"),
					presentCommand("disable-ssh"),
					presentCommand("ssh-enabled"),
					presentCommand("ssh"),
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
					presentCommand("update-service"),
					presentCommand("delete-service"),
					presentCommand("rename-service"),
				}, {
					presentCommand("create-service-key"),
					presentCommand("service-keys"),
					presentCommand("service-key"),
					presentCommand("delete-service-key"),
				}, {
					presentCommand("bind-service"),
					presentCommand("unbind-service"),
				}, {
					presentCommand("bind-route-service"),
					presentCommand("unbind-route-service"),
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
				}, {
					presentCommand("allow-space-ssh"),
					presentCommand("disallow-space-ssh"),
					presentCommand("space-ssh-allowed"),
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
				{
					presentCommand("router-groups"),
				},
			},
		}, {
			Name: T("ROUTES"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentCommand("routes"),
					presentCommand("create-route"),
					presentCommand("check-route"),
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
				{
					presentCommand("share-private-domain"),
					presentCommand("unshare-private-domain"),
				},
			},
		}, {
			Name: T("SPACE ADMIN"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentCommand("space-quotas"),
					presentCommand("space-quota"),
					presentCommand("create-space-quota"),
					presentCommand("update-space-quota"),
					presentCommand("delete-space-quota"),
					presentCommand("set-space-quota"),
					presentCommand("unset-space-quota"),
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
					presentCommand("purge-service-instance"),
				}, {
					presentCommand("service-access"),
					presentCommand("enable-service-access"),
					presentCommand("disable-service-access"),
				},
			},
		}, {
			Name: T("SECURITY GROUP"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentCommand("security-group"),
					presentCommand("security-groups"),
					presentCommand("create-security-group"),
					presentCommand("update-security-group"),
					presentCommand("delete-security-group"),
					presentCommand("bind-security-group"),
					presentCommand("unbind-security-group"),
				}, {
					presentCommand("bind-staging-security-group"),
					presentCommand("staging-security-groups"),
					presentCommand("unbind-staging-security-group"),
				}, {
					presentCommand("bind-running-security-group"),
					presentCommand("running-security-groups"),
					presentCommand("unbind-running-security-group"),
				},
			},
		}, {
			Name: T("ENVIRONMENT VARIABLE GROUPS"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentCommand("running-environment-variable-group"),
					presentCommand("staging-environment-variable-group"),
					presentCommand("set-staging-environment-variable-group"),
					presentCommand("set-running-environment-variable-group"),
				},
			},
		},
		{
			Name: T("FEATURE FLAGS"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentCommand("feature-flags"),
					presentCommand("feature-flag"),
					presentCommand("enable-feature-flag"),
					presentCommand("disable-feature-flag"),
				},
			},
		}, {
			Name: T("ADVANCED"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentCommand("curl"),
					presentCommand("config"),
					presentCommand("oauth-token"),
					presentCommand("ssh-code"),
				},
			},
		}, {
			Name: T("ADD/REMOVE PLUGIN REPOSITORY"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentCommand("add-plugin-repo"),
					presentCommand("remove-plugin-repo"),
					presentCommand("list-plugin-repos"),
					presentCommand("repo-plugins"),
				},
			},
		}, {
			Name: T("ADD/REMOVE PLUGIN"),
			CommandSubGroups: [][]cmdPresenter{
				{
					presentCommand("plugins"),
					presentCommand("install-plugin"),
					presentCommand("uninstall-plugin"),
				},
			},
		}, {
			Name: T("INSTALLED PLUGIN COMMANDS"),
			CommandSubGroups: [][]cmdPresenter{
				presentPluginCommands(),
			},
		},
	}

	return presenter
}

func (p appPresenter) Title(name string) string {
	return terminal.HeaderColor(name)
}

func (c groupedCommands) SubTitle(name string) string {
	return terminal.HeaderColor(name + ":")
}

func maxPluginCommandNameLength(plugins map[string]pluginconfig.PluginMetadata, maxNameLen int) int {
	for _, pluginMetadata := range plugins {
		for _, cmd := range pluginMetadata.Commands {
			if nameLen := utf8.RuneCountInString(cmd.Name); nameLen > maxNameLen {
				maxNameLen = nameLen
			}
		}
	}

	return maxNameLen
}
