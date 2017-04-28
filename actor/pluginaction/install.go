package pluginaction

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/gofileutils/fileutils"
)

//go:generate counterfeiter . PluginMetadata

type PluginMetadata interface {
	GetMetadata(pluginPath string) (configv3.Plugin, error)
}

//go:generate counterfeiter . CommandList

type CommandList interface {
	HasCommand(string) bool
	HasAlias(string) bool
}

// PluginInvalidError is returned with a plugin is invalid because it is
// missing a name or has 0 commands.
type PluginInvalidError struct {
	Path string
}

func (e PluginInvalidError) Error() string {
	return "File {{.Path}} is not a valid cf CLI plugin binary."
}

// PluginCommandConflictError is returned when a plugin command name conflicts
// with a core or existing plugin command name.
type PluginCommandsConflictError struct {
	PluginName     string
	PluginVersion  string
	CommandAliases []string
	CommandNames   []string
}

func (e PluginCommandsConflictError) Error() string {
	return ""
}

// FileExists returns true if the file exists. It returns false if the file
// doesn't exist or there is an error checking.
func (actor Actor) FileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

func (actor Actor) IsPluginInstalled(pluginName string) bool {
	_, isInstalled := actor.config.GetPlugin(pluginName)
	return isInstalled
}

func (actor Actor) GetAndValidatePlugin(pluginMetadata PluginMetadata, commandList CommandList, path string) (configv3.Plugin, error) {
	plugin, err := pluginMetadata.GetMetadata(path)
	if err != nil {
		return configv3.Plugin{}, err
	}

	if plugin.Name == "" || len(plugin.Commands) == 0 {
		return configv3.Plugin{}, PluginInvalidError{Path: path}
	}

	installedPlugins := actor.config.Plugins()

	conflictingNames := []string{}
	conflictingAliases := []string{}

	for _, command := range plugin.Commands {
		if commandList.HasCommand(command.Name) || commandList.HasAlias(command.Name) {
			conflictingNames = append(conflictingNames, command.Name)
		}

		if commandList.HasAlias(command.Alias) || commandList.HasCommand(command.Alias) {
			conflictingAliases = append(conflictingAliases, command.Alias)
		}

		for _, installedPlugin := range installedPlugins {
			// we do not error if a plugins commands conflict with previous
			// versions of the same plugin
			if plugin.Name == installedPlugin.Name {
				continue
			}

			for _, installedCommand := range installedPlugin.Commands {
				if command.Name == installedCommand.Name || command.Name == installedCommand.Alias {
					conflictingNames = append(conflictingNames, command.Name)
				}

				if command.Alias != "" &&
					(command.Alias == installedCommand.Alias || command.Alias == installedCommand.Name) {
					conflictingAliases = append(conflictingAliases, command.Alias)
				}
			}
		}
	}

	sort.Slice(conflictingNames, func(i, j int) bool {
		return strings.ToLower(conflictingNames[i]) < strings.ToLower(conflictingNames[j])
	})

	sort.Slice(conflictingAliases, func(i, j int) bool {
		return strings.ToLower(conflictingAliases[i]) < strings.ToLower(conflictingAliases[j])
	})

	if len(conflictingNames) > 0 || len(conflictingAliases) > 0 {
		return configv3.Plugin{}, PluginCommandsConflictError{
			PluginName:     plugin.Name,
			PluginVersion:  plugin.Version.String(),
			CommandNames:   conflictingNames,
			CommandAliases: conflictingAliases,
		}
	}

	return plugin, nil
}

func (actor Actor) InstallPluginFromPath(path string, plugin configv3.Plugin) error {
	installPath := filepath.Join(actor.config.PluginHome(), filepath.Base(path))
	err := fileutils.CopyPathToPath(path, installPath)
	if err != nil {
		return err
	}

	plugin.Location = installPath

	actor.config.AddPlugin(plugin)

	err = actor.config.WritePluginConfig()
	if err != nil {
		return err
	}

	return nil
}
