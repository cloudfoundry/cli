package pluginaction

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/plugin"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/generic"
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

// CreateExecutableCopy makes a temporary copy of a plugin binary and makes it
// executable.
//
// config.PluginHome() + /temp is used as the temp dir instead of the system
// temp for security reasons.
func (actor Actor) CreateExecutableCopy(path string, tempPluginDir string) (string, error) {
	tempFile, err := makeTempFile(tempPluginDir)
	if err != nil {
		return "", err
	}

	// add '.exe' to the temp file if on Windows
	executablePath := generic.ExecutableFilename(tempFile.Name())
	err = os.Rename(tempFile.Name(), executablePath)
	if err != nil {
		return "", err
	}

	err = fileutils.CopyPathToPath(path, executablePath)
	if err != nil {
		return "", err
	}

	err = os.Chmod(executablePath, 0700)
	if err != nil {
		return "", err
	}

	return executablePath, nil
}

// DownloadExecutableBinaryFromURL fetches a plugin binary from the specified
// URL, if it exists.
func (actor Actor) DownloadExecutableBinaryFromURL(pluginURL string, tempPluginDir string, proxyReader plugin.ProxyReader) (string, error) {
	tempFile, err := makeTempFile(tempPluginDir)
	if err != nil {
		return "", err
	}

	err = actor.client.DownloadPlugin(pluginURL, tempFile.Name(), proxyReader)
	if err != nil {
		return "", err
	}

	return tempFile.Name(), nil
}

// FileExists returns true if the file exists. It returns false if the file
// doesn't exist or there is an error checking.
func (actor Actor) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (actor Actor) GetAndValidatePlugin(pluginMetadata PluginMetadata, commandList CommandList, path string) (configv3.Plugin, error) {
	plugin, err := pluginMetadata.GetMetadata(path)
	if err != nil || plugin.Name == "" || len(plugin.Commands) == 0 {
		return configv3.Plugin{}, actionerror.PluginInvalidError{Err: err}
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

	if len(conflictingNames) > 0 || len(conflictingAliases) > 0 {
		sort.Slice(conflictingNames, func(i, j int) bool {
			return strings.ToLower(conflictingNames[i]) < strings.ToLower(conflictingNames[j])
		})

		sort.Slice(conflictingAliases, func(i, j int) bool {
			return strings.ToLower(conflictingAliases[i]) < strings.ToLower(conflictingAliases[j])
		})

		return configv3.Plugin{}, actionerror.PluginCommandsConflictError{
			PluginName:     plugin.Name,
			PluginVersion:  plugin.Version.String(),
			CommandNames:   conflictingNames,
			CommandAliases: conflictingAliases,
		}
	}

	return plugin, nil
}

func (actor Actor) InstallPluginFromPath(path string, plugin configv3.Plugin) error {
	installPath := generic.ExecutableFilename(filepath.Join(actor.config.PluginHome(), plugin.Name))
	err := fileutils.CopyPathToPath(path, installPath)
	if err != nil {
		return err
	}
	// rwxr-xr-x so that multiple users can share the same $CF_PLUGIN_HOME
	err = os.Chmod(installPath, 0755)
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

func makeTempFile(tempDir string) (*os.File, error) {
	tempFile, err := ioutil.TempFile(tempDir, "")
	if err != nil {
		return nil, err
	}

	err = tempFile.Close()
	if err != nil {
		return nil, err
	}

	return tempFile, nil
}
