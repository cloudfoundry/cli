package pluginaction

import (
	"fmt"
	"runtime"

	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/generic"
)

type PluginInfo struct {
	Name     string
	Version  string
	URL      string
	Checksum string
}

// PluginNotFoundError is an error returned when a plugin is not found.
type PluginNotFoundInRepositoryError struct {
	PluginName     string
	RepositoryName string
}

// Error outputs a plugin not found error message.
func (e PluginNotFoundInRepositoryError) Error() string {
	return fmt.Sprintf("Plugin %s not found in repository %s", e.PluginName, e.RepositoryName)
}

// PluginNotFoundError is an error returned when a plugin is not found.
type PluginNotFoundInAnyRepositoryError struct {
	PluginName string
}

// Error outputs a plugin not found error message.
func (e PluginNotFoundInAnyRepositoryError) Error() string {
	return fmt.Sprintf("Plugin %s not found in any registered repo", e.PluginName)
}

func (actor Actor) GetPluginInfoFromRepositoryForPlatform(pluginName string, pluginRepo configv3.PluginRepository, platform string) (PluginInfo, error) {
	pluginRepository, err := actor.client.GetPluginRepository(pluginRepo.URL)
	if err != nil {
		return PluginInfo{}, err
	}

	for _, plugin := range pluginRepository.Plugins {
		if plugin.Name == pluginName {
			for _, pluginBinary := range plugin.Binaries {
				if pluginBinary.Platform == platform {
					return PluginInfo{Name: plugin.Name, Version: plugin.Version, URL: pluginBinary.URL, Checksum: pluginBinary.Checksum}, nil
				}
			}
		}
	}

	return PluginInfo{}, PluginNotFoundInRepositoryError{PluginName: pluginName, RepositoryName: pluginRepo.Name}
}

// GetPluginInfoFromRepositoriesForPlatform returns the newest version of the specified plugin
// and all the repositories that contain that version.
func (actor Actor) GetPluginInfoFromRepositoriesForPlatform(pluginName string, pluginRepos []configv3.PluginRepository, platform string) (PluginInfo, []string, error) {
	var reposWithPlugin []string
	var newestPluginInfo PluginInfo

	for _, repo := range pluginRepos {
		pluginInfo, err := actor.GetPluginInfoFromRepositoryForPlatform(pluginName, repo, platform)
		if _, ok := err.(PluginNotFoundInRepositoryError); ok {
			continue
		} else if err != nil {
			return PluginInfo{}, nil, err
		}
		if len(reposWithPlugin) == 0 || lessThan(newestPluginInfo.Version, pluginInfo.Version) {
			newestPluginInfo = pluginInfo
			reposWithPlugin = []string{repo.Name}
		} else if pluginInfo.Version == newestPluginInfo.Version {
			reposWithPlugin = append(reposWithPlugin, repo.Name)
		}
	}

	if len(reposWithPlugin) == 0 {
		return PluginInfo{}, nil, PluginNotFoundInAnyRepositoryError{PluginName: pluginName}
	}
	return newestPluginInfo, reposWithPlugin, nil
}

// GetPlatformString exists solely for the purposes of mocking it out for command-layers tests.
func (actor Actor) GetPlatformString(runtimeGOOS string, runtimeGOARCH string) string {
	return generic.GeneratePlatform(runtime.GOOS, runtime.GOARCH)
}
