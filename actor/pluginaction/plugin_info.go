package pluginaction

import (
	"runtime"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/generic"
)

type PluginInfo struct {
	Name     string
	Version  string
	URL      string
	Checksum string
}

// GetPluginInfoFromRepositoriesForPlatform returns the newest version of the specified plugin
// and all the repositories that contain that version.
func (actor Actor) GetPluginInfoFromRepositoriesForPlatform(pluginName string, pluginRepos []configv3.PluginRepository, platform string) (PluginInfo, []string, error) {
	var reposWithPlugin []string
	var newestPluginInfo PluginInfo
	var pluginFoundWithIncompatibleBinary bool

	for _, repo := range pluginRepos {
		pluginInfo, err := actor.getPluginInfoFromRepositoryForPlatform(pluginName, repo, platform)
		switch err.(type) {
		case actionerror.PluginNotFoundInRepositoryError:
			continue
		case actionerror.NoCompatibleBinaryError:
			pluginFoundWithIncompatibleBinary = true
			continue
		case nil:
			if len(reposWithPlugin) == 0 || lessThan(newestPluginInfo.Version, pluginInfo.Version) {
				newestPluginInfo = pluginInfo
				reposWithPlugin = []string{repo.Name}
			} else if pluginInfo.Version == newestPluginInfo.Version {
				reposWithPlugin = append(reposWithPlugin, repo.Name)
			}
		default:
			return PluginInfo{}, nil, actionerror.FetchingPluginInfoFromRepositoryError{
				RepositoryName: repo.Name,
				Err:            err,
			}
		}
	}

	if len(reposWithPlugin) == 0 {
		if pluginFoundWithIncompatibleBinary {
			return PluginInfo{}, nil, actionerror.NoCompatibleBinaryError{}
		}
		return PluginInfo{}, nil, actionerror.PluginNotFoundInAnyRepositoryError{PluginName: pluginName}
	}
	return newestPluginInfo, reposWithPlugin, nil
}

// GetPlatformString exists solely for the purposes of mocking it out for command-layers tests.
func (actor Actor) GetPlatformString(runtimeGOOS string, runtimeGOARCH string) string {
	return generic.GeneratePlatform(runtime.GOOS, runtime.GOARCH)
}

// getPluginInfoFromRepositoryForPlatform returns the plugin info, if found, from
// the specified repository for the specified platform.
func (actor Actor) getPluginInfoFromRepositoryForPlatform(pluginName string, pluginRepo configv3.PluginRepository, platform string) (PluginInfo, error) {
	pluginRepository, err := actor.client.GetPluginRepository(pluginRepo.URL)
	if err != nil {
		return PluginInfo{}, err
	}

	var pluginFoundWithIncompatibleBinary bool

	for _, plugin := range pluginRepository.Plugins {
		if plugin.Name == pluginName {
			for _, pluginBinary := range plugin.Binaries {
				if pluginBinary.Platform == platform {
					return PluginInfo{
						Name:     plugin.Name,
						Version:  plugin.Version,
						URL:      pluginBinary.URL,
						Checksum: pluginBinary.Checksum,
					}, nil
				}
			}
			pluginFoundWithIncompatibleBinary = true
		}
	}

	if pluginFoundWithIncompatibleBinary {
		return PluginInfo{}, actionerror.NoCompatibleBinaryError{}
	}

	return PluginInfo{}, actionerror.PluginNotFoundInRepositoryError{
		PluginName:     pluginName,
		RepositoryName: pluginRepo.Name,
	}
}
