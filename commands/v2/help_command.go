package v2

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/actors/v2actions"
	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
	"code.cloudfoundry.org/cli/commands/ui"
	"code.cloudfoundry.org/cli/utils/config"
)

type helpCategory struct {
	categoryName string
	commandList  []string
}

const (
	BLANKLINE = ""
	CF_NAME   = "cf"
)

var helpCategoryList = []helpCategory{
	{
		categoryName: "GETTING STARTED:",
		commandList:  []string{"help", "version", "login", "logout", "passwd", "target", BLANKLINE, "api", "auth"},
	},
	{
		categoryName: "APPS:",
		commandList:  []string{"apps", "app", BLANKLINE, "push", "scale", "delete", "rename", BLANKLINE, "start", "stop", "restart", "restage", "restart-app-instance", BLANKLINE, "events", "files", "logs", BLANKLINE, "env", "set-env", "unset-env", BLANKLINE, "stacks", "stack", BLANKLINE, "copy-source", "create-app-manifest", BLANKLINE, "get-health-check", "set-health-check", "enable-ssh", "disable-ssh", "ssh-enabled", "ssh"},
	},
	{
		categoryName: "SERVICES:",
		commandList:  []string{"marketplace", "services", "service", BLANKLINE, "create-service", "update-service", "delete-service", "rename-service", BLANKLINE, "create-service-key", "service-keys", "service-key", "delete-service-key", BLANKLINE, "bind-service", "unbind-service", BLANKLINE, "bind-route-service", "unbind-route-service", BLANKLINE, "create-user-provided-service", "update-user-provided-service"},
	},
	{
		categoryName: "ORGS:",
		commandList:  []string{"orgs", "org", BLANKLINE, "create-org", "delete-org", "rename-org"},
	},
	{
		categoryName: "SPACES:",
		commandList:  []string{"spaces", "space", BLANKLINE, "create-space", "delete-space", "rename-space", BLANKLINE, "allow-space-ssh", "disallow-space-ssh", "space-ssh-allowed"},
	},
	{
		categoryName: "DOMAINS:",
		commandList:  []string{"domains", "create-domain", "delete-domain", "create-shared-domain", "delete-shared-domain", BLANKLINE, "router-groups"},
	},
	{
		categoryName: "ROUTES:",
		commandList:  []string{"routes", "create-route", "check-route", "map-route", "unmap-route", "delete-route", "delete-orphaned-routes"},
	},
	{
		categoryName: "BUILDPACKS:",
		commandList:  []string{"buildpacks", "create-buildpack", "update-buildpack", "rename-buildpack", "delete-buildpack"},
	},
	{
		categoryName: "USER ADMIN:",
		commandList:  []string{"create-user", "delete-user", BLANKLINE, "org-users", "set-org-role", "unset-org-role", BLANKLINE, "space-users", "set-space-role", "unset-space-role"},
	},
	{
		categoryName: "ORG ADMIN:",
		commandList:  []string{"quotas", "quota", "set-quota", BLANKLINE, "create-quota", "delete-quota", "update-quota", BLANKLINE, "share-private-domain", "unshare-private-domain"},
	},
	{
		categoryName: "SPACE ADMIN:",
		commandList:  []string{"space-quotas", "space-quota", BLANKLINE, "create-space-quota", "update-space-quota", "delete-space-quota", BLANKLINE, "set-space-quota", "unset-space-quota"},
	},
	{
		categoryName: "SERVICE ADMIN:",
		commandList:  []string{"service-auth-tokens", "create-service-auth-token", "update-service-auth-token", "delete-service-auth-token", BLANKLINE, "service-brokers", "create-service-broker", "update-service-broker", "delete-service-broker", "rename-service-broker", BLANKLINE, "migrate-service-instances", "purge-service-offering", "purge-service-instance", BLANKLINE, "service-access", "enable-service-access", "disable-service-access"},
	},
	{
		categoryName: "SECURITY GROUP:",
		commandList:  []string{"security-group", "security-groups", "create-security-group", "update-security-group", "delete-security-group", "bind-security-group", "unbind-security-group", BLANKLINE, "bind-staging-security-group", "staging-security-groups", "unbind-staging-security-group", BLANKLINE, "bind-running-security-group", "running-security-groups", "unbind-running-security-group"},
	},
	{
		categoryName: "ENVIRONMENT VARIABLE GROUPS:",
		commandList:  []string{"running-environment-variable-group", "staging-environment-variable-group", "set-staging-environment-variable-group", "set-running-environment-variable-group"},
	},
	{
		categoryName: "FEATURE FLAGS:",
		commandList:  []string{"feature-flags", "feature-flag", "enable-feature-flag", "disable-feature-flag"},
	},
	{
		categoryName: "ADVANCED:",
		commandList:  []string{"curl", "config", "oauth-token", "ssh-code"},
	},
	{
		categoryName: "ADD/REMOVE PLUGIN REPOSITORY:",
		commandList:  []string{"add-plugin-repo", "remove-plugin-repo", "list-plugin-repos", "repo-plugins"},
	},
	{
		categoryName: "ADD/REMOVE PLUGIN:",
		commandList:  []string{"plugins", "install-plugin", "uninstall-plugin"},
	},
}

//go:generate counterfeiter . HelpActor

// HelpActor handles the business logic of the help command
type HelpActor interface {
	// GetCommandInfo returns back a help command information for the given
	// command
	GetCommandInfo(interface{}, string) (v2actions.CommandInfo, error)

	// GetAllNamesAndDescriptions returns a list of all commands
	GetAllNamesAndDescriptions(interface{}) map[string]v2actions.CommandInfo
}

type HelpCommand struct {
	UI     UI
	Actor  HelpActor
	Config commands.Config

	OptionalArgs flags.CommandName `positional-args:"yes"`
	usage        interface{}       `usage:"CF_NAME help [COMMAND]"`
}

func (cmd *HelpCommand) Setup(config commands.Config) error {
	cmd.UI = ui.NewUI(config)
	cmd.Actor = v2actions.NewActor()
	cmd.Config = config
	return nil
}

func (cmd HelpCommand) Execute(args []string) error {
	var err error
	if cmd.OptionalArgs.CommandName == "" {
		cmd.displayFullHelp()
	} else {
		err = cmd.displayCommand()
	}

	return err
}

func (cmd HelpCommand) displayFullHelp() {
	cmd.displayHelpPreamble()
	cmd.displayAllCommands()
	cmd.displayHelpFooter()
}

func (cmd HelpCommand) displayHelpPreamble() {
	cmd.UI.DisplayHelpHeader("NAME:")
	cmd.UI.DisplayTextWithKeyTranslations("   {{.CommandName}} - {{.CommandDescription}}",
		[]string{"CommandDescription"},
		map[string]interface{}{
			"CommandName":        CF_NAME,
			"CommandDescription": "A command line tool to interact with Cloud Foundry",
		})
	cmd.UI.DisplayNewline()

	cmd.UI.DisplayHelpHeader("USAGE:")
	cmd.UI.DisplayTextWithKeyTranslations("   {{.CommandName}} {{.CommandUsage}}",
		[]string{"CommandUsage"},
		map[string]interface{}{
			"CommandName":  CF_NAME,
			"CommandUsage": "[global options] command [arguments...] [command options]",
		})
	cmd.UI.DisplayNewline()

	cmd.UI.DisplayHelpHeader("VERSION:")
	cmd.UI.DisplayText("   {{.Version}}-{{.Time}}", map[string]interface{}{
		"Version": cf.Version,
		"Time":    cf.BuiltOnDate,
	})
	cmd.UI.DisplayNewline()
}

func (cmd HelpCommand) displayAllCommands() {
	cmdInfo := cmd.Actor.GetAllNamesAndDescriptions(Commands)
	pluginCommands := []config.PluginCommand{}

	for _, pluginCommand := range cmd.Config.PluginConfig() {
		pluginCommands = append(pluginCommands, pluginCommand.Commands...)
	}
	longestCmd := cmd.longestCommandName(cmdInfo, pluginCommands)

	for _, category := range helpCategoryList {
		cmd.UI.DisplayHelpHeader(category.categoryName)

		for _, command := range category.commandList {
			if command == BLANKLINE {
				cmd.UI.DisplayNewline()
				continue
			}

			cmd.UI.DisplayText("   {{.CommandName}}{{.Gap}}{{.CommandDescription}}", map[string]interface{}{
				"CommandName":        cmdInfo[command].Name,
				"CommandDescription": cmdInfo[command].Description,
				"Gap":                strings.Repeat(" ", longestCmd+1-len(command)),
			})
		}

		cmd.UI.DisplayNewline()

		cmd.UI.DisplayHelpHeader("INSTALLED PLUGIN COMMANDS:")
		for _, pluginCommand := range pluginCommands {
			cmd.UI.DisplayText("   {{.CommandName}}{{.Gap}}{{.CommandDescription}}", map[string]interface{}{
				"CommandName":        pluginCommand.Name,
				"CommandDescription": pluginCommand.HelpText,
				"Gap":                strings.Repeat(" ", longestCmd+1-len(pluginCommand.Name)),
			})
		}
		cmd.UI.DisplayNewline()
	}
}

func (_ HelpCommand) longestCommandName(cmds map[string]v2actions.CommandInfo, pluginCmds []config.PluginCommand) int {
	longest := 0
	for name, _ := range cmds {
		if len(name) > longest {
			longest = len(name)
		}
	}
	for _, command := range pluginCmds {
		if len(command.Name) > longest {
			longest = len(command.Name)
		}
	}
	return longest
}

func (cmd HelpCommand) displayHelpFooter() {
	cmd.UI.DisplayHelpHeader("ENVIRONMENT VARIABLES:")
	cmd.UI.DisplayTextWithKeyTranslations("   {{.ENVName}}                     {{.Description}}",
		[]string{"Description"},
		map[string]interface{}{
			"ENVName":     "CF_COLOR=false",
			"Description": "Do not colorize output",
		})
	cmd.UI.DisplayTextWithKeyTranslations("   {{.ENVName}}               {{.Description}}",
		[]string{"Description"},
		map[string]interface{}{
			"ENVName":     "CF_HOME=path/to/dir/",
			"Description": "Override path to default config directory",
		})
	cmd.UI.DisplayTextWithKeyTranslations("   {{.ENVName}}        {{.Description}}",
		[]string{"Description"},
		map[string]interface{}{
			"ENVName":     "CF_PLUGIN_HOME=path/to/dir/",
			"Description": "Override path to default plugin config director",
		})
	cmd.UI.DisplayTextWithKeyTranslations("   {{.ENVName}}              {{.Description}}",
		[]string{"Description"},
		map[string]interface{}{
			"ENVName":     "CF_STAGING_TIMEOUT=15",
			"Description": "Max wait time for buildpack staging, in minutes",
		})
	cmd.UI.DisplayTextWithKeyTranslations("   {{.ENVName}}               {{.Description}}",
		[]string{"Description"},
		map[string]interface{}{
			"ENVName":     "CF_STARTUP_TIMEOUT=5",
			"Description": "Max wait time for app instance startup, in minutes",
		})
	cmd.UI.DisplayTextWithKeyTranslations("   {{.ENVName}}                      {{.Description}}",
		[]string{"Description"},
		map[string]interface{}{
			"ENVName":     "CF_TRACE=true",
			"Description": "Print API request diagnostics to stdout",
		})
	cmd.UI.DisplayTextWithKeyTranslations("   {{.ENVName}}         {{.Description}}",
		[]string{"Description"},
		map[string]interface{}{
			"ENVName":     "CF_TRACE=path/to/trace.log",
			"Description": "Append API request diagnostics to a log file",
		})
	cmd.UI.DisplayTextWithKeyTranslations("   {{.ENVName}} {{.Description}}",
		[]string{"Description"},
		map[string]interface{}{
			"ENVName":     "https_proxy=proxy.example.com:8080",
			"Description": "Enable HTTP proxying for API requests",
		})
	cmd.UI.DisplayNewline()

	cmd.UI.DisplayHelpHeader("GLOBAL OPTIONS:")
	cmd.UI.DisplayTextWithKeyTranslations("   {{.ENVName}}                         {{.Description}}",
		[]string{"Description"},
		map[string]interface{}{
			"ENVName":     "--help, -h",
			"Description": "Show help",
		})
	cmd.UI.DisplayTextWithKeyTranslations("   {{.ENVName}}                                 {{.Description}}",
		[]string{"Description"},
		map[string]interface{}{
			"ENVName":     "-v",
			"Description": "Print API request diagnostics to stdout",
		})
}

func (cmd HelpCommand) displayCommand() error {
	cmdInfo, err := cmd.Actor.GetCommandInfo(Commands, cmd.OptionalArgs.CommandName)
	if err != nil {
		return err
	}

	cmd.UI.DisplayText("NAME:")
	cmd.UI.DisplayTextWithKeyTranslations("    {{.CommandName}} - {{.CommandDescription}}",
		[]string{"CommandDescription"},
		map[string]interface{}{
			"CommandName":        cmdInfo.Name,
			"CommandDescription": cmdInfo.Description,
		})
	cmd.UI.DisplayText("")

	//TODO: Figure out the best way to dynamically determine this
	usageString := strings.Replace(cmdInfo.Usage, "CF_NAME", CF_NAME, -1)
	cmd.UI.DisplayText("USAGE:")
	cmd.UI.DisplayTextWithKeyTranslations("    {{.CommandUsage}}",
		[]string{"CommandUsage"},
		map[string]interface{}{
			"CommandUsage": usageString,
		})
	cmd.UI.DisplayText("")

	if cmdInfo.Alias != "" {
		cmd.UI.DisplayText("ALIAS:")
		cmd.UI.DisplayText("    {{.Alias}}",
			map[string]interface{}{
				"Alias": cmdInfo.Alias,
			})
		cmd.UI.DisplayText("")
	}

	if len(cmdInfo.Flags) != 0 {
		cmd.UI.DisplayText("OPTIONS:")
		nameWidth := cmd.longestFlagWidth(cmdInfo.Flags) + 6
		for _, flag := range cmdInfo.Flags {
			var name string
			if flag.Short != "" && flag.Long != "" {
				name = fmt.Sprintf("--%s, -%s", flag.Long, flag.Short)
			} else if flag.Short != "" {
				name = "-" + flag.Short
			} else {
				name = "--" + flag.Long
			}

			cmd.UI.DisplayTextWithKeyTranslations("    {{.Flags}}{{.Spaces}}{{.Description}}",
				[]string{"Description"},
				map[string]interface{}{
					"Flags":       name,
					"Spaces":      strings.Repeat(" ", nameWidth-len(name)),
					"Description": flag.Description,
				})
		}
	}

	return nil
}

func (_ HelpCommand) longestFlagWidth(flags []v2actions.CommandFlag) int {
	longest := 0
	for _, flag := range flags {
		var name string
		if flag.Short != "" && flag.Long != "" {
			name = fmt.Sprintf("--%s, -%s", flag.Long, flag.Short)
		} else if flag.Short != "" {
			name = "-" + flag.Short
		} else {
			name = "--" + flag.Long
		}
		if len(name) > longest {
			longest = len(name)
		}
	}
	return longest
}
