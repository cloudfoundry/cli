package internal

import (
	"fmt"
	"sort"
	"strings"

	"code.cloudfoundry.org/cli/actors/v2actions"
	"code.cloudfoundry.org/cli/utils/config"
	"code.cloudfoundry.org/cli/utils/sortutils"
)

type HelpCategory struct {
	CategoryName string
	CommandList  []string
}

const BLANKLINE = ""

var HelpCategoryList = []HelpCategory{
	{
		CategoryName: "GETTING STARTED:",
		CommandList:  []string{"help", "version", "login", "logout", "passwd", "target", BLANKLINE, "api", "auth"},
	},
	{
		CategoryName: "APPS:",
		CommandList:  []string{"apps", "app", BLANKLINE, "push", "scale", "delete", "rename", BLANKLINE, "start", "stop", "restart", "restage", "restart-app-instance", BLANKLINE, "events", "files", "logs", BLANKLINE, "env", "set-env", "unset-env", BLANKLINE, "stacks", "stack", BLANKLINE, "copy-source", "create-app-manifest", BLANKLINE, "get-health-check", "set-health-check", "enable-ssh", "disable-ssh", "ssh-enabled", "ssh"},
	},
	{
		CategoryName: "SERVICES:",
		CommandList:  []string{"marketplace", "services", "service", BLANKLINE, "create-service", "update-service", "delete-service", "rename-service", BLANKLINE, "create-service-key", "service-keys", "service-key", "delete-service-key", BLANKLINE, "bind-service", "unbind-service", BLANKLINE, "bind-route-service", "unbind-route-service", BLANKLINE, "create-user-provided-service", "update-user-provided-service"},
	},
	{
		CategoryName: "ORGS:",
		CommandList:  []string{"orgs", "org", BLANKLINE, "create-org", "delete-org", "rename-org"},
	},
	{
		CategoryName: "SPACES:",
		CommandList:  []string{"spaces", "space", BLANKLINE, "create-space", "delete-space", "rename-space", BLANKLINE, "allow-space-ssh", "disallow-space-ssh", "space-ssh-allowed"},
	},
	{
		CategoryName: "DOMAINS:",
		CommandList:  []string{"domains", "create-domain", "delete-domain", "create-shared-domain", "delete-shared-domain", BLANKLINE, "router-groups"},
	},
	{
		CategoryName: "ROUTES:",
		CommandList:  []string{"routes", "create-route", "check-route", "map-route", "unmap-route", "delete-route", "delete-orphaned-routes"},
	},
	{
		CategoryName: "BUILDPACKS:",
		CommandList:  []string{"buildpacks", "create-buildpack", "update-buildpack", "rename-buildpack", "delete-buildpack"},
	},
	{
		CategoryName: "USER ADMIN:",
		CommandList:  []string{"create-user", "delete-user", BLANKLINE, "org-users", "set-org-role", "unset-org-role", BLANKLINE, "space-users", "set-space-role", "unset-space-role"},
	},
	{
		CategoryName: "ORG ADMIN:",
		CommandList:  []string{"quotas", "quota", "set-quota", BLANKLINE, "create-quota", "delete-quota", "update-quota", BLANKLINE, "share-private-domain", "unshare-private-domain"},
	},
	{
		CategoryName: "SPACE ADMIN:",
		CommandList:  []string{"space-quotas", "space-quota", BLANKLINE, "create-space-quota", "update-space-quota", "delete-space-quota", BLANKLINE, "set-space-quota", "unset-space-quota"},
	},
	{
		CategoryName: "SERVICE ADMIN:",
		CommandList:  []string{"service-auth-tokens", "create-service-auth-token", "update-service-auth-token", "delete-service-auth-token", BLANKLINE, "service-brokers", "create-service-broker", "update-service-broker", "delete-service-broker", "rename-service-broker", BLANKLINE, "migrate-service-instances", "purge-service-offering", "purge-service-instance", BLANKLINE, "service-access", "enable-service-access", "disable-service-access"},
	},
	{
		CategoryName: "SECURITY GROUP:",
		CommandList:  []string{"security-group", "security-groups", "create-security-group", "update-security-group", "delete-security-group", "bind-security-group", "unbind-security-group", BLANKLINE, "bind-staging-security-group", "staging-security-groups", "unbind-staging-security-group", BLANKLINE, "bind-running-security-group", "running-security-groups", "unbind-running-security-group"},
	},
	{
		CategoryName: "ENVIRONMENT VARIABLE GROUPS:",
		CommandList:  []string{"running-environment-variable-group", "staging-environment-variable-group", "set-staging-environment-variable-group", "set-running-environment-variable-group"},
	},
	{
		CategoryName: "FEATURE FLAGS:",
		CommandList:  []string{"feature-flags", "feature-flag", "enable-feature-flag", "disable-feature-flag"},
	},
	{
		CategoryName: "ADVANCED:",
		CommandList:  []string{"curl", "config", "oauth-token", "ssh-code"},
	},
	{
		CategoryName: "ADD/REMOVE PLUGIN REPOSITORY:",
		CommandList:  []string{"add-plugin-repo", "remove-plugin-repo", "list-plugin-repos", "repo-plugins"},
	},
	{
		CategoryName: "ADD/REMOVE PLUGIN:",
		CommandList:  []string{"plugins", "install-plugin", "uninstall-plugin"},
	},
}

func ConvertPluginToCommandInfo(plugin config.PluginCommand) v2actions.CommandInfo {
	commandInfo := v2actions.CommandInfo{
		Name:        plugin.Name,
		Description: plugin.HelpText,
		Alias:       plugin.Alias,
		Usage:       plugin.UsageDetails.Usage,
		Flags:       []v2actions.CommandFlag{},
	}

	flagNames := sortutils.Alphabetic{}
	for flag := range plugin.UsageDetails.Options {
		flagNames = append(flagNames, flag)
	}
	sort.Sort(flagNames)

	for _, flag := range flagNames {
		description := plugin.UsageDetails.Options[flag]
		strippedFlag := strings.Trim(flag, "-")
		switch len(flag) {
		case 1:
			commandInfo.Flags = append(commandInfo.Flags,
				v2actions.CommandFlag{
					Short:       strippedFlag,
					Description: description,
				})
		default:
			commandInfo.Flags = append(commandInfo.Flags,
				v2actions.CommandFlag{
					Long:        strippedFlag,
					Description: description,
				})
		}
	}

	return commandInfo
}

func LongestCommandName(cmds map[string]v2actions.CommandInfo, pluginCmds []config.PluginCommand) int {
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

func LongestFlagWidth(flags []v2actions.CommandFlag) int {
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
