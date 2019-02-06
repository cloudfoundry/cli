// +build V7

package common

import (
	"reflect"

	"code.cloudfoundry.org/cli/command/plugin"
	"code.cloudfoundry.org/cli/command/v6"
	"code.cloudfoundry.org/cli/plugin/rpc"
)

// var Commands commandList
var FallbackCommands V2CommandList

type V2CommandList struct {
	V2App v6.V3AppCommand `command:"app" description:"Display health and status for an app"`
}

var Commands commandList

type commandList struct {
	rpc.CommandList
	Help             HelpCommand                    `command:"help" alias:"h" description:"Show help"`
	AddPluginRepo    plugin.AddPluginRepoCommand    `command:"add-plugin-repo" description:"Add a new plugin repository"`
	InstallPlugin    InstallPluginCommand           `command:"install-plugin" description:"Install CLI plugin"`
	ListPluginRepos  plugin.ListPluginReposCommand  `command:"list-plugin-repos" description:"List all the added plugin repositories"`
	Plugins          plugin.PluginsCommand          `command:"plugins" description:"List commands of installed plugins"`
	RemovePluginRepo plugin.RemovePluginRepoCommand `command:"remove-plugin-repo" description:"Remove a plugin repository"`
	RepoPlugins      plugin.RepoPluginsCommand      `command:"repo-plugins" description:"List all available plugins in specified repository or in all added repositories"`
	UninstallPlugin  plugin.UninstallPluginCommand  `command:"uninstall-plugin" description:"Uninstall CLI plugin"`
}

// HasCommand returns true if the command name is in the command list.
func (c commandList) HasCommand(name string) bool {
	if name == "" {
		return false
	}

	cType := reflect.TypeOf(c)
	_, found := cType.FieldByNameFunc(
		func(fieldName string) bool {
			field, _ := cType.FieldByName(fieldName)
			return field.Tag.Get("command") == name
		},
	)

	return found
}

// HasAlias returns true if the command alias is in the command list.
func (c commandList) HasAlias(alias string) bool {
	if alias == "" {
		return false
	}

	cType := reflect.TypeOf(c)
	_, found := cType.FieldByNameFunc(
		func(fieldName string) bool {
			field, _ := cType.FieldByName(fieldName)
			return field.Tag.Get("alias") == alias
		},
	)

	return found
}
