package pluginaction

import (
	"fmt"

	"github.com/blang/semver"
)

type OutdatedPlugin struct {
	Name           string
	CurrentVersion string
	LatestVersion  string
}

func (actor Actor) GetOutdatedPlugins() ([]OutdatedPlugin, error) {
	var outdatedPlugins []OutdatedPlugin

	repoPlugins := map[string]string{}
	for _, repo := range actor.config.PluginRepositories() {
		repository, err := actor.client.GetPluginRepository(repo.URL)
		if err != nil {
			return nil, GettingPluginRepositoryError{Name: repo.Name, Message: err.Error()}
		}

		for _, plugin := range repository.Plugins {
			existingVersion, exist := repoPlugins[plugin.Name]
			if exist {
				if lessThan(existingVersion, plugin.Version) {
					repoPlugins[plugin.Name] = plugin.Version
				}
			} else {
				repoPlugins[plugin.Name] = plugin.Version
			}
		}
	}

	for _, installedPlugin := range actor.config.Plugins() {
		repoVersion, exist := repoPlugins[installedPlugin.Name]
		if exist && lessThan(installedPlugin.Version.String(), repoVersion) {
			outdatedPlugins = append(outdatedPlugins, OutdatedPlugin{
				Name:           installedPlugin.Name,
				CurrentVersion: installedPlugin.Version.String(),
				LatestVersion:  repoVersion,
			})
		}
	}

	return outdatedPlugins, nil
}

func lessThan(version1 string, version2 string) bool {
	v1, err := semver.Make(version1)
	if err != nil {
		return false
	}

	v2, err := semver.Make(version2)
	if err != nil {
		return false
	}

	return v1.LT(v2)
}

// GettingPluginRepositoryError is returned when there's an error
// accessing the plugin repository
type GettingPluginRepositoryError struct {
	Name    string
	Message string
}

func (e GettingPluginRepositoryError) Error() string {
	return fmt.Sprintf("Could not get plugin repository '%s': %s", e.Name, e.Message)
}

// PluginRepositoryNotFoundError is returned when the plugin Repository was not found.
type PluginRepositoryNotFoundError struct {
	Name string
}

func (e PluginRepositoryNotFoundError) Error() string {
	return fmt.Sprintf("Plugin repository '%s' was not found.", e.Name)
}

// NoPluginRepositoriesError is returned when no plugin repositories
// are found.
type NoPluginRepositoriesError struct {
}

func (e NoPluginRepositoriesError) Error() string {
	return "No plugin repositories added."
}
